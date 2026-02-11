package grpc

import (
	"context"
	"errors"
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	taskspb "task-manager-microservice/proto/tasks"
	"task-manager-microservice/tasks/core"
)

type Server struct {
	taskspb.UnimplementedCategoriesServiceServer
	taskspb.UnimplementedTasksServiceServer

	log     *slog.Logger
	service *core.Service
}

func NewServer(log *slog.Logger, service *core.Service) *Server {
	return &Server{log: log, service: service}
}

func (s *Server) Ping(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	if err := s.service.Ping(ctx); err != nil {
		s.log.Error("ping failed", "error", err)
		return nil, status.Error(codes.Internal, "ping failed")
	}
	return &emptypb.Empty{}, nil
}

// Categories

func (s *Server) CreateCategory(ctx context.Context, req *taskspb.CreateCategoryRequest) (*taskspb.Category, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	c, err := s.service.CreateCategory(ctx, req.GetName())
	if err != nil {
		return nil, s.mapErr(err)
	}

	return categoryToPB(c), nil
}

func (s *Server) GetCategory(ctx context.Context, req *taskspb.GetCategoryRequest) (*taskspb.Category, error) {
	if req == nil || req.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}

	c, err := s.service.GetCategory(ctx, req.GetId())
	if err != nil {
		return nil, s.mapErr(err)
	}

	return categoryToPB(c), nil
}

func (s *Server) ListCategories(ctx context.Context, _ *taskspb.ListCategoriesRequest) (*taskspb.ListCategoriesResponse, error) {
	items, err := s.service.ListCategories(ctx)
	if err != nil {
		return nil, s.mapErr(err)
	}

	out := make([]*taskspb.Category, 0, len(items))
	for _, c := range items {
		out = append(out, categoryToPB(c))
	}

	return &taskspb.ListCategoriesResponse{Categories: out}, nil
}

func (s *Server) UpdateCategory(ctx context.Context, req *taskspb.UpdateCategoryRequest) (*taskspb.Category, error) {
	if req == nil || req.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}

	c, err := s.service.UpdateCategory(ctx, req.GetId(), req.GetName())
	if err != nil {
		return nil, s.mapErr(err)
	}

	return categoryToPB(c), nil
}

func (s *Server) DeleteCategory(ctx context.Context, req *taskspb.DeleteCategoryRequest) (*emptypb.Empty, error) {
	if req == nil || req.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}

	if err := s.service.DeleteCategory(ctx, req.GetId()); err != nil {
		return nil, s.mapErr(err)
	}

	return &emptypb.Empty{}, nil
}

// Tasks

func (s *Server) CreateTask(ctx context.Context, req *taskspb.CreateTaskRequest) (*taskspb.Task, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.GetCategoryId() < 0 {
		return nil, status.Error(codes.InvalidArgument, "category_id cannot be negative")
	}

	var catID *int64
	if req.GetCategoryId() != 0 {
		id := req.GetCategoryId()
		catID = &id
	}

	t, err := s.service.CreateTask(ctx, catID, req.GetName(), req.GetDescription())
	if err != nil {
		return nil, s.mapErr(err)
	}

	return taskToPB(t), nil
}

func (s *Server) GetTask(ctx context.Context, req *taskspb.GetTaskRequest) (*taskspb.Task, error) {
	if req == nil || req.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}

	t, err := s.service.GetTask(ctx, req.GetId())
	if err != nil {
		return nil, s.mapErr(err)
	}

	return taskToPB(t), nil
}

func (s *Server) ListTask(ctx context.Context, req *taskspb.ListTaskRequest) (*taskspb.ListTaskResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	var f core.ListTasksFilter

	// status_filter oneof
	switch x := req.StatusFilter.(type) {
	case *taskspb.ListTaskRequest_Status:
		st, err := pbStatusToCore(x.Status)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid status")
		}
		f.Status = &st
	}

	// category_filter oneof
	switch x := req.CategoryFilter.(type) {
	case *taskspb.ListTaskRequest_CategoryId:
		id := x.CategoryId
		f.CategoryID = &id
	case *taskspb.ListTaskRequest_WithoutCategory:
		f.WithoutCategory = x.WithoutCategory
	}

	f.Limit = int(req.GetLimit())
	f.Offset = int(req.GetOffset())

	items, err := s.service.ListTasks(ctx, f)
	if err != nil {
		return nil, s.mapErr(err)
	}

	out := make([]*taskspb.Task, 0, len(items))
	for _, t := range items {
		out = append(out, taskToPB(t))
	}

	return &taskspb.ListTaskResponse{Tasks: out}, nil
}

func (s *Server) UpdateTask(ctx context.Context, req *taskspb.UpdateTaskRequest) (*taskspb.Task, error) {
	if req == nil || req.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}
	if req.GetCategoryId() < 0 {
		return nil, status.Error(codes.InvalidArgument, "category_id cannot be negative")
	}

	st, err := pbStatusToCore(req.GetStatus())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid status")
	}

	var catID *int64
	if req.GetCategoryId() != 0 {
		id := req.GetCategoryId()
		catID = &id
	}

	t := core.Task{
		ID:          req.GetId(),
		CategoryID:  catID, // nil => снять категорию
		Name:        req.GetName(),
		Description: req.GetDescription(),
		Status:      st,
	}

	updated, err := s.service.UpdateTask(ctx, t)
	if err != nil {
		return nil, s.mapErr(err)
	}

	return taskToPB(updated), nil
}

func (s *Server) DeleteTask(ctx context.Context, req *taskspb.DeleteTaskRequest) (*emptypb.Empty, error) {
	if req == nil || req.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}

	if err := s.service.DeleteTask(ctx, req.GetId()); err != nil {
		return nil, s.mapErr(err)
	}

	return &emptypb.Empty{}, nil
}

// Helpers

func categoryToPB(c core.Category) *taskspb.Category {
	return &taskspb.Category{
		Id:        c.ID,
		Name:      c.Name,
		CreatedAt: timestamppb.New(c.CreatedAt),
	}
}

func taskToPB(t core.Task) *taskspb.Task {
	var catID int64
	if t.CategoryID != nil {
		catID = *t.CategoryID
	}

	return &taskspb.Task{
		Id:          t.ID,
		CategoryId:  catID, // 0 => без категории
		Name:        t.Name,
		Description: t.Description,
		Status:      coreStatusToPB(t.Status),
		CreatedAt:   timestamppb.New(t.CreatedAt),
		UpdatedAt:   timestamppb.New(t.UpdatedAt),
	}
}

func pbStatusToCore(st taskspb.TaskStatus) (core.TaskStatus, error) {
	switch st {
	case taskspb.TaskStatus_TASK_STATUS_TODO:
		return core.TODO, nil
	case taskspb.TaskStatus_TASK_STATUS_IN_PROGRESS:
		return core.InProgress, nil
	case taskspb.TaskStatus_TASK_STATUS_DONE:
		return core.Done, nil
	case taskspb.TaskStatus_TASK_STATUS_ARCHIVED:
		return core.Archived, nil
	default:
		return core.TODO, errors.New("unknown status")
	}
}

func coreStatusToPB(st core.TaskStatus) taskspb.TaskStatus {
	switch st {
	case core.TODO:
		return taskspb.TaskStatus_TASK_STATUS_TODO
	case core.InProgress:
		return taskspb.TaskStatus_TASK_STATUS_IN_PROGRESS
	case core.Done:
		return taskspb.TaskStatus_TASK_STATUS_DONE
	case core.Archived:
		return taskspb.TaskStatus_TASK_STATUS_ARCHIVED
	default:
		return taskspb.TaskStatus_TASK_STATUS_TODO
	}
}

func (s *Server) mapErr(err error) error {
	switch {
	// categories
	case errors.Is(err, core.ErrCategoryInvalidArgs):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, core.ErrCategoryNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, core.ErrCategoryAlreadyExists):
		return status.Error(codes.AlreadyExists, err.Error())

	// tasks
	case errors.Is(err, core.ErrTaskInvalidArgs):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, core.ErrTaskNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, core.ErrTaskAlreadyExists):
		return status.Error(codes.AlreadyExists, err.Error())

	default:
		s.log.Error("internal error", "error", err)
		return status.Error(codes.Internal, "internal error")
	}
}
