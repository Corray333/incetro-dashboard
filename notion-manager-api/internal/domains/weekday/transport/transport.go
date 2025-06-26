package transport

import (
	"context"

	"github.com/Corray333/employee_dashboard/internal/domains/feedback/entities/feedback"
	feedbackpb "github.com/Corray333/employee_dashboard/internal/domains/feedback/transport/grpc"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type service interface {
	ListFeedbacks(ctx context.Context, filter *feedback.Feedback) ([]feedback.Feedback, error)
}
type WeekdayTransport struct {
	feedbackpb.UnimplementedFeedbackServiceServer
	service service
}

func NewWeekdayTransport(grpcServer *grpc.Server, service service) *WeekdayTransport {
	t := &WeekdayTransport{
		service: service,
	}
	feedbackpb.RegisterFeedbackServiceServer(grpcServer, t)

	return t
}

func convertProtoToDomain(p *feedbackpb.Feedback) *feedback.Feedback {
	createdDate := p.GetCreatedDate().AsTime().UTC()
	id, _ := uuid.Parse(p.GetId().GetValue())
	taskID, _ := uuid.Parse(p.GetTaskId().GetValue())
	projectID, _ := uuid.Parse(p.GetProjectId().GetValue())

	return &feedback.Feedback{
		ID:          id,
		Text:        p.GetText(),
		Type:        p.GetType(),
		Priority:    p.GetPriority(),
		TaskID:      taskID,
		ProjectID:   projectID,
		CreatedDate: createdDate,
		Direction:   p.GetDirection(),
		Status:      p.GetStatus(),
	}
}

func convertDomainToProto(f *feedback.Feedback) *feedbackpb.Feedback {
	ts := timestamppb.New(f.CreatedDate)

	return &feedbackpb.Feedback{
		Id:          &feedbackpb.UUID{Value: f.ID.String()},
		Text:        f.Text,
		Type:        f.Type,
		Priority:    f.Priority,
		TaskId:      &feedbackpb.UUID{Value: f.TaskID.String()},
		ProjectId:   &feedbackpb.UUID{Value: f.ProjectID.String()},
		CreatedDate: ts,
		Direction:   f.Direction,
		Status:      f.Status,
	}
}

func (s *WeekdayTransport) ListFeedbacks(ctx context.Context, req *feedbackpb.ListFeedbacksRequest) (*feedbackpb.ListFeedbacksResponse, error) {
	// Преобразовать req.Filter из proto в domain модель
	filter := convertProtoToDomain(req.GetFilter())

	// Получить список из базы
	feedbacks, err := s.service.ListFeedbacks(ctx, filter)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list feedbacks: %v", err)
	}

	// Преобразовать результат обратно в proto
	var protoFeedbacks []*feedbackpb.Feedback
	for _, f := range feedbacks {
		protoFeedbacks = append(protoFeedbacks, convertDomainToProto(&f))
	}

	return &feedbackpb.ListFeedbacksResponse{
		Feedbacks: protoFeedbacks,
	}, nil
}
