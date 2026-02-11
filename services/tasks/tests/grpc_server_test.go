package tests

import (
	"context"
	"io"
	"log/slog"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	taskspb "task-manager-microservice/proto/tasks"
	taskgrpc "task-manager-microservice/tasks/adapters/grpc"
	"task-manager-microservice/tasks/core"
)

const bufConnSize = 1024 * 1024

func assertStatusCode(t *testing.T, err error, wantCode codes.Code) {
	t.Helper()

	if err == nil {
		t.Fatalf("expected error with code %v, got nil", wantCode)
	}
	if status.Code(err) != wantCode {
		t.Fatalf("expected %v, got %v (%v)", wantCode, status.Code(err), err)
	}
}

func strPtr(v string) *string {
	return &v
}

func int64Ptr(v int64) *int64 {
	return &v
}

func newTasksGRPCClient(t *testing.T, db core.DB) (taskspb.TasksServiceClient, func()) {
	t.Helper()

	listener := bufconn.Listen(bufConnSize)
	grpcServer := grpc.NewServer()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	service := core.NewService(db)
	tasksServer := taskgrpc.NewServer(logger, service)
	taskspb.RegisterTasksServiceServer(grpcServer, tasksServer)

	serveErr := make(chan error, 1)
	go func() {
		serveErr <- grpcServer.Serve(listener)
	}()

	dialCtx, cancelDial := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelDial()

	conn, err := grpc.DialContext(
		dialCtx,
		"bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return listener.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		grpcServer.Stop()
		listener.Close()
		t.Fatalf("failed to dial bufconn server: %v", err)
	}

	cleanup := func() {
		_ = conn.Close()
		grpcServer.Stop()
		_ = listener.Close()
		select {
		case <-serveErr:
		default:
		}
	}

	return taskspb.NewTasksServiceClient(conn), cleanup
}

func TestGRPCUpdateTask_InvalidCases(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		build    func(taskID int64) *taskspb.UpdateTaskRequest
		wantCode codes.Code
	}{
		{
			name: "update_mask_name_without_name",
			build: func(taskID int64) *taskspb.UpdateTaskRequest {
				return &taskspb.UpdateTaskRequest{
					Id: taskID,
					UpdateMask: &fieldmaskpb.FieldMask{
						Paths: []string{"name"},
					},
				}
			},
			wantCode: codes.InvalidArgument,
		},
		{
			name: "update_mask_unknown_field",
			build: func(taskID int64) *taskspb.UpdateTaskRequest {
				return &taskspb.UpdateTaskRequest{
					Id: taskID,
					UpdateMask: &fieldmaskpb.FieldMask{
						Paths: []string{"unknown"},
					},
				}
			},
			wantCode: codes.InvalidArgument,
		},
		{
			name: "without_fields",
			build: func(taskID int64) *taskspb.UpdateTaskRequest {
				return &taskspb.UpdateTaskRequest{Id: taskID}
			},
			wantCode: codes.InvalidArgument,
		},
		{
			name: "category_not_found",
			build: func(taskID int64) *taskspb.UpdateTaskRequest {
				return &taskspb.UpdateTaskRequest{
					Id:         taskID,
					CategoryId: int64Ptr(999),
				}
			},
			wantCode: codes.NotFound,
		},
		{
			name: "negative_category_id",
			build: func(taskID int64) *taskspb.UpdateTaskRequest {
				return &taskspb.UpdateTaskRequest{
					Id:         taskID,
					CategoryId: int64Ptr(-1),
				}
			},
			wantCode: codes.InvalidArgument,
		},
		{
			name: "update_mask_category_id_without_value",
			build: func(taskID int64) *taskspb.UpdateTaskRequest {
				return &taskspb.UpdateTaskRequest{
					Id: taskID,
					UpdateMask: &fieldmaskpb.FieldMask{
						Paths: []string{"category_id"},
					},
				}
			},
			wantCode: codes.InvalidArgument,
		},
		{
			name: "invalid_id",
			build: func(int64) *taskspb.UpdateTaskRequest {
				return &taskspb.UpdateTaskRequest{Id: 0}
			},
			wantCode: codes.InvalidArgument,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db := newFakeDB()
			task, err := db.CreateTask(context.Background(), nil, "old name", "old description")
			if err != nil {
				t.Fatalf("failed to prepare task: %v", err)
			}

			client, cleanup := newTasksGRPCClient(t, db)
			defer cleanup()

			_, err = client.UpdateTask(context.Background(), tc.build(task.ID))
			assertStatusCode(t, err, tc.wantCode)
		})
	}
}

func TestGRPCUpdateTask_UpdateMaskRestrictsUpdates(t *testing.T) {
	t.Parallel()

	db := newFakeDB()
	task, err := db.CreateTask(context.Background(), nil, "old name", "old description")
	if err != nil {
		t.Fatalf("failed to prepare task: %v", err)
	}

	client, cleanup := newTasksGRPCClient(t, db)
	defer cleanup()

	updated, err := client.UpdateTask(context.Background(), &taskspb.UpdateTaskRequest{
		Id:          task.ID,
		Name:        strPtr("new name"),
		Description: strPtr("new description should be ignored"),
		UpdateMask: &fieldmaskpb.FieldMask{
			Paths: []string{"name"},
		},
	})
	if err != nil {
		t.Fatalf("UpdateTask returned error: %v", err)
	}
	if updated.Name != "new name" {
		t.Fatalf("expected name %q, got %q", "new name", updated.GetName())
	}
	if updated.Description != "old description" {
		t.Fatalf("expected description %q, got %q", "old description", updated.GetDescription())
	}

	stored, err := db.GetTask(context.Background(), task.ID)
	if err != nil {
		t.Fatalf("failed to load task from db: %v", err)
	}
	if stored.Description != "old description" {
		t.Fatalf("expected stored description %q, got %q", "old description", stored.Description)
	}
}

func TestGRPCUpdateTask_RemoveCategory_WithCategoryIdZero(t *testing.T) {
	t.Parallel()

	db := newFakeDB()
	category, err := db.CreateCategory(context.Background(), "work")
	if err != nil {
		t.Fatalf("failed to prepare category: %v", err)
	}

	categoryID := category.ID
	task, err := db.CreateTask(context.Background(), &categoryID, "task", "description")
	if err != nil {
		t.Fatalf("failed to prepare task: %v", err)
	}

	client, cleanup := newTasksGRPCClient(t, db)
	defer cleanup()

	updated, err := client.UpdateTask(context.Background(), &taskspb.UpdateTaskRequest{
		Id:         task.ID,
		CategoryId: int64Ptr(0),
	})
	if err != nil {
		t.Fatalf("UpdateTask returned error: %v", err)
	}
	if updated.GetCategoryId() != 0 {
		t.Fatalf("expected category_id 0, got %d", updated.GetCategoryId())
	}
	stored, err := db.GetTask(context.Background(), task.ID)
	if err != nil {
		t.Fatalf("failed to load task from db: %v", err)
	}
	if stored.CategoryID != nil {
		t.Fatalf("expected category to be removed, got %d", *stored.CategoryID)
	}
}
