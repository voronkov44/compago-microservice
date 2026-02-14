package tasks

import (
	"context"
	"fmt"
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	apicore "task-manager-microservice/api/core"
	taskspb "task-manager-microservice/proto/tasks"
)

type Client struct {
	log  *slog.Logger
	conn *grpc.ClientConn

	tasks taskspb.TasksServiceClient
	cats  taskspb.CategoriesServiceClient
}

func NewClient(address string, log *slog.Logger) (*Client, error) {
	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("new grpc client for %s: %w", address, err)
	}

	return &Client{
		log:   log,
		conn:  conn,
		tasks: taskspb.NewTasksServiceClient(conn),
		cats:  taskspb.NewCategoriesServiceClient(conn),
	}, nil
}

func (c *Client) Close() error { return c.conn.Close() }

// ---- Pinger

func (c *Client) Ping(ctx context.Context) error {
	_, err := c.tasks.Ping(ctx, &emptypb.Empty{})
	return mapGRPCErr(err)
}

// ---- Categories

func (c *Client) CreateCategory(ctx context.Context, name string) (apicore.Category, error) {
	resp, err := c.cats.CreateCategory(ctx, &taskspb.CreateCategoryRequest{Name: name})
	if err != nil {
		return apicore.Category{}, mapGRPCErr(err)
	}
	return catFromPB(resp), nil
}

func (c *Client) GetCategory(ctx context.Context, id int64) (apicore.Category, error) {
	resp, err := c.cats.GetCategory(ctx, &taskspb.GetCategoryRequest{Id: id})
	if err != nil {
		return apicore.Category{}, mapGRPCErr(err)
	}
	return catFromPB(resp), nil
}

func (c *Client) ListCategories(ctx context.Context) ([]apicore.Category, error) {
	resp, err := c.cats.ListCategories(ctx, &taskspb.ListCategoriesRequest{})
	if err != nil {
		return nil, mapGRPCErr(err)
	}

	out := make([]apicore.Category, 0, len(resp.GetCategories()))
	for _, it := range resp.GetCategories() {
		out = append(out, catFromPB(it))
	}
	return out, nil
}

func (c *Client) UpdateCategory(ctx context.Context, id int64, name string) (apicore.Category, error) {
	resp, err := c.cats.UpdateCategory(ctx, &taskspb.UpdateCategoryRequest{Id: id, Name: name})
	if err != nil {
		return apicore.Category{}, mapGRPCErr(err)
	}
	return catFromPB(resp), nil
}

func (c *Client) DeleteCategory(ctx context.Context, id int64) error {
	_, err := c.cats.DeleteCategory(ctx, &taskspb.DeleteCategoryRequest{Id: id})
	return mapGRPCErr(err)
}

// ---- Tasks

func (c *Client) CreateTask(ctx context.Context, categoryID *int64, name, description string) (apicore.Task, error) {
	var cid int64
	if categoryID != nil {
		cid = *categoryID
	}

	resp, err := c.tasks.CreateTask(ctx, &taskspb.CreateTaskRequest{
		CategoryId:  cid,
		Name:        name,
		Description: description,
	})
	if err != nil {
		return apicore.Task{}, mapGRPCErr(err)
	}
	return taskFromPB(resp), nil
}

func (c *Client) GetTask(ctx context.Context, id int64) (apicore.Task, error) {
	resp, err := c.tasks.GetTask(ctx, &taskspb.GetTaskRequest{Id: id})
	if err != nil {
		return apicore.Task{}, mapGRPCErr(err)
	}
	return taskFromPB(resp), nil
}

func (c *Client) ListTasks(ctx context.Context, f apicore.ListTasksFilter) ([]apicore.Task, error) {
	req := &taskspb.ListTaskRequest{
		Limit:  int32(f.Limit),
		Offset: int32(f.Offset),
	}

	if f.Status != nil {
		pb, err := statusCoreToPB(*f.Status)
		if err != nil {
			return nil, apicore.ErrBadArguments
		}
		req.StatusFilter = &taskspb.ListTaskRequest_Status{Status: pb}
	}

	if f.CategoryID != nil {
		req.CategoryFilter = &taskspb.ListTaskRequest_CategoryId{CategoryId: *f.CategoryID}
	} else if f.WithoutCategory {
		req.CategoryFilter = &taskspb.ListTaskRequest_WithoutCategory{WithoutCategory: true}
	}

	resp, err := c.tasks.ListTask(ctx, req)
	if err != nil {
		return nil, mapGRPCErr(err)
	}

	out := make([]apicore.Task, 0, len(resp.GetTasks()))
	for _, it := range resp.GetTasks() {
		out = append(out, taskFromPB(it))
	}
	return out, nil
}

func (c *Client) PatchTask(ctx context.Context, id int64, p apicore.TaskPatch) (apicore.Task, error) {
	req := &taskspb.UpdateTaskRequest{Id: id}
	mask := &fieldmaskpb.FieldMask{}

	if p.CategoryID != nil {
		v := *p.CategoryID
		req.CategoryId = &v
		mask.Paths = append(mask.Paths, "category_id")
	}
	if p.Name != nil {
		v := *p.Name
		req.Name = &v
		mask.Paths = append(mask.Paths, "name")
	}
	if p.Description != nil {
		v := *p.Description
		req.Description = &v
		mask.Paths = append(mask.Paths, "description")
	}
	if p.Status != nil {
		pb, err := statusCoreToPB(*p.Status)
		if err != nil {
			return apicore.Task{}, apicore.ErrBadArguments
		}
		req.Status = &pb
		mask.Paths = append(mask.Paths, "status")
	}

	if len(mask.Paths) == 0 {
		return apicore.Task{}, apicore.ErrBadArguments
	}
	req.UpdateMask = mask

	resp, err := c.tasks.UpdateTask(ctx, req)
	if err != nil {
		return apicore.Task{}, mapGRPCErr(err)
	}
	return taskFromPB(resp), nil
}

func (c *Client) DeleteTask(ctx context.Context, id int64) error {
	_, err := c.tasks.DeleteTask(ctx, &taskspb.DeleteTaskRequest{Id: id})
	return mapGRPCErr(err)
}

var _ apicore.Tasks = (*Client)(nil)

// ---- helpers

func mapGRPCErr(err error) error {
	if err == nil {
		return nil
	}
	switch status.Code(err) {
	case codes.InvalidArgument:
		return apicore.ErrBadArguments
	case codes.NotFound:
		return apicore.ErrNotFound
	case codes.AlreadyExists:
		return apicore.ErrAlreadyExists
	case codes.Unavailable, codes.DeadlineExceeded, codes.Canceled:
		return apicore.ErrUnavailable
	default:
		return err
	}
}

func statusPBToCore(st taskspb.TaskStatus) apicore.TaskStatus {
	switch st {
	case taskspb.TaskStatus_TASK_STATUS_TODO:
		return apicore.StatusTODO
	case taskspb.TaskStatus_TASK_STATUS_IN_PROGRESS:
		return apicore.StatusInProgress
	case taskspb.TaskStatus_TASK_STATUS_DONE:
		return apicore.StatusDone
	case taskspb.TaskStatus_TASK_STATUS_ARCHIVED:
		return apicore.StatusArchived
	default:
		return apicore.StatusTODO
	}
}

func statusCoreToPB(st apicore.TaskStatus) (taskspb.TaskStatus, error) {
	switch st {
	case apicore.StatusTODO:
		return taskspb.TaskStatus_TASK_STATUS_TODO, nil
	case apicore.StatusInProgress:
		return taskspb.TaskStatus_TASK_STATUS_IN_PROGRESS, nil
	case apicore.StatusDone:
		return taskspb.TaskStatus_TASK_STATUS_DONE, nil
	case apicore.StatusArchived:
		return taskspb.TaskStatus_TASK_STATUS_ARCHIVED, nil
	default:
		return taskspb.TaskStatus_TASK_STATUS_TODO, apicore.ErrBadArguments
	}
}

func catFromPB(x *taskspb.Category) apicore.Category {
	return apicore.Category{
		ID:        x.GetId(),
		Name:      x.GetName(),
		CreatedAt: x.GetCreatedAt().AsTime(),
	}
}

func taskFromPB(x *taskspb.Task) apicore.Task {
	var cid *int64
	if x.GetCategoryId() != 0 {
		v := x.GetCategoryId()
		cid = &v
	}
	return apicore.Task{
		ID:          x.GetId(),
		CategoryID:  cid,
		Name:        x.GetName(),
		Description: x.GetDescription(),
		Status:      statusPBToCore(x.GetStatus()),
		CreatedAt:   x.GetCreatedAt().AsTime(),
		UpdatedAt:   x.GetUpdatedAt().AsTime(),
	}
}
