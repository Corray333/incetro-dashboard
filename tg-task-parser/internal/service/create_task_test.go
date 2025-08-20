package service_test

import (
	"context"
	"testing"

	"github.com/corray333/tg-task-parser/internal/entities/task"
	"github.com/corray333/tg-task-parser/internal/repositories/yatracker"
	svc "github.com/corray333/tg-task-parser/internal/service"
	"github.com/google/uuid"
)

type mockProjectGetter struct{ id uuid.UUID; err error }

func (m mockProjectGetter) GetProjectByChatID(_ context.Context, _ int64) (uuid.UUID, error) { return m.id, m.err }

type mockTaskCreator struct{ pageID string; err error }

func (m mockTaskCreator) CreateTask(_ context.Context, _ *task.Task, _ uuid.UUID) (string, error) {
	return m.pageID, m.err
}

type mockYaRepo struct {
	searchIssues []yatracker.Issue
	searchErr    error
	createIssue  *yatracker.Issue
	createErr    error
}

func (m mockYaRepo) CreateTask(_ context.Context, _ *task.Task) (*yatracker.Issue, error) { return m.createIssue, m.createErr }
func (m mockYaRepo) SearchTasksByName(_ context.Context, _ *task.Task) ([]yatracker.Issue, error) {
	return m.searchIssues, m.searchErr
}

func TestServiceCreateTask_GeneralProject(t *testing.T) {
	projectID := uuid.New()
	s := svc.New(
		svc.WithProjectByChatIDGetter(mockProjectGetter{id: projectID}),
		svc.WithTaskCreator(mockTaskCreator{pageID: "11111111-2222-3333-4444-555555555555"}),
	)

	text, err := s.CreateTask(context.Background(), 123, "Создать #задача", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "Задача создана: https://notion.so/11111111222233334444555555555555"
	if text != want {
		t.Fatalf("unexpected text: got %q want %q", text, want)
	}
}

func TestServiceCreateTask_Tracker_Found(t *testing.T) {
	trackerProjectID, _ := uuid.Parse("183dd045-c92a-42c3-83ba-5030fbb3451f")
	ya := mockYaRepo{searchIssues: []yatracker.Issue{{Key: "INC-42", Summary: "Fix bug"}}}
	s := svc.New(
		svc.WithProjectByChatIDGetter(mockProjectGetter{id: trackerProjectID}),
		svc.WithTaskCreator(mockTaskCreator{pageID: "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"}),
		svc.WithYaTrackerRepo(ya),
	)

	text, err := s.CreateTask(context.Background(), 1, "починить #задача", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "Задача \"INC-42: Fix bug\" создана:\n\n- Яндекс.Трекер: https://tracker.yandex.ru/INC-42\n- Notion: https://notion.so/aaaaaaaaBBBBccccddddeeeeeeeeeeee"
	// Note: service removes dashes only; case remains. Adjust expected to reflect dash removal precisely.
	want = "Задача \"INC-42: Fix bug\" создана:\n\n- Яндекс.Трекер: https://tracker.yandex.ru/INC-42\n- Notion: https://notion.so/aaaaaaaabbbbccccddddeeeeeeeeeeee"
	if text != want {
		t.Fatalf("unexpected text: got %q want %q", text, want)
	}
}

func TestServiceCreateTask_Tracker_Create(t *testing.T) {
	trackerProjectID, _ := uuid.Parse("183dd045-c92a-42c3-83ba-5030fbb3451f")
	ya := mockYaRepo{createIssue: &yatracker.Issue{Key: "INC-43", Summary: "New task"}}
	s := svc.New(
		svc.WithProjectByChatIDGetter(mockProjectGetter{id: trackerProjectID}),
		svc.WithTaskCreator(mockTaskCreator{pageID: "00000000-0000-0000-0000-000000000000"}),
		svc.WithYaTrackerRepo(ya),
	)

	text, err := s.CreateTask(context.Background(), 1, "сделать #задача", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "Задача \"INC-43: New task\" создана:\n\n- Яндекс.Трекер: https://tracker.yandex.ru/INC-43\n- Notion: https://notion.so/00000000000000000000000000000000"
	if text != want {
		t.Fatalf("unexpected text: got %q want %q", text, want)
	}
}

func TestServiceCreateTask_NoTaskTag_ReturnsEmpty(t *testing.T) {
	projectID := uuid.New()
	s := svc.New(
		svc.WithProjectByChatIDGetter(mockProjectGetter{id: projectID}),
		svc.WithTaskCreator(mockTaskCreator{pageID: "deadbeef-dead-beef-dead-beefdeadbeef"}),
	)

	text, err := s.CreateTask(context.Background(), 1, "обычное сообщение без тега", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if text != "" {
		t.Fatalf("expected empty text, got %q", text)
	}
}

func TestServiceCreateTask_Errors(t *testing.T) {
	projectErr := mockProjectGetter{err: context.DeadlineExceeded}
	s := svc.New(svc.WithProjectByChatIDGetter(projectErr))
	if _, err := s.CreateTask(context.Background(), 1, "создать #задача", ""); err == nil {
		t.Fatalf("expected error from project getter")
	}

	projectID := uuid.New()
	s2 := svc.New(
		svc.WithProjectByChatIDGetter(mockProjectGetter{id: projectID}),
		svc.WithTaskCreator(mockTaskCreator{err: context.Canceled}),
	)
	if _, err := s2.CreateTask(context.Background(), 1, "создать #задача", ""); err == nil {
		t.Fatalf("expected error from task creator")
	}

	trackerProjectID, _ := uuid.Parse("183dd045-c92a-42c3-83ba-5030fbb3451f")
	yaErr := mockYaRepo{searchErr: context.DeadlineExceeded}
	s3 := svc.New(
		svc.WithProjectByChatIDGetter(mockProjectGetter{id: trackerProjectID}),
		svc.WithTaskCreator(mockTaskCreator{pageID: "00000000-0000-0000-0000-000000000000"}),
		svc.WithYaTrackerRepo(yaErr),
	)
	if _, err := s3.CreateTask(context.Background(), 1, "создать #задача", ""); err == nil {
		t.Fatalf("expected error from yaTracker search")
	}

	yaCreateErr := mockYaRepo{searchIssues: []yatracker.Issue{}, createErr: context.Canceled}
	s4 := svc.New(
		svc.WithProjectByChatIDGetter(mockProjectGetter{id: trackerProjectID}),
		svc.WithTaskCreator(mockTaskCreator{pageID: "00000000-0000-0000-0000-000000000000"}),
		svc.WithYaTrackerRepo(yaCreateErr),
	)
	if _, err := s4.CreateTask(context.Background(), 1, "создать #задача", ""); err == nil {
		t.Fatalf("expected error from yaTracker create")
	}
}