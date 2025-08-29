package service

import (
	"context"

	"github.com/corray333/tg-task-parser/internal/entities/employee"
	"github.com/corray333/tg-task-parser/internal/entities/task"
	"github.com/google/uuid"
)

type baseService interface {
	taskCreator
	projectsGetter
	feedbackLister
}

type repository interface {
	chatToProjectLinker
	projectByChatIDGetter
	messageMetaSetter
	messageMetaScanner
	tgMessageSaver
	employeeTgIDUpdater
	employeeTgIDByIDGetter
}

type employeeTgIDByIDGetter interface {
	GetEmployeeByProfileID(ctx context.Context, profileID uuid.UUID) (*employee.Employee, error)
}

type notionRepo interface {
	feedbackAnswerer
	feedbackCreator
	topicsGetter
	employeesWithIncorrectTimeGetter
}

type yaTrackerRepo interface {
	yaTrackerTaskCreator
	yaTrackerTaskSearcher
}

type taskMsgParser interface {
	ParseMessage(ctx context.Context, message string) (*task.Task, error)
}

type Service struct {
	taskCreator           taskCreator
	projectsGetter        projectsGetter
	chatToProjectLinker   chatToProjectLinker
	projectByChatIDGetter projectByChatIDGetter
	feedbackLister        feedbackLister
	messageMetaSetter     messageMetaSetter
	messageMetaScanner    messageMetaScanner

	feedbackAnswerer                feedbackAnswerer
	feedbackCreator                 feedbackCreator
	tgMessageSaver                  tgMessageSaver
	topicsGetter                    topicsGetter
	employeeTgIDUpdater             employeeTgIDUpdater
	employeeTgIDByIDGetter          employeeTgIDByIDGetter
	incorrectTimeNotificationSender incorrectTimeNotificationSender

	yaTrackerTaskCreator  yaTrackerTaskCreator
	yaTrackerTaskSearcher yaTrackerTaskSearcher

	repository repository
	notionRepo notionRepo
	tgRepo     tgMessageSender

	taskMsgParser taskMsgParser
}

type option func(*Service)

func New(options ...option) *Service {
	s := &Service{}
	for _, opt := range options {
		opt(s)
	}
	return s
}

func WithTaskMsgParser(taskMsgParser taskMsgParser) option {
	return func(s *Service) {
		s.taskMsgParser = taskMsgParser
	}
}

func WithYaTrackerRepo(yaTrackerRepo yaTrackerRepo) option {
	return func(s *Service) {
		s.yaTrackerTaskCreator = yaTrackerRepo
		s.yaTrackerTaskSearcher = yaTrackerRepo
	}
}

func WithProjectByChatIDGetter(projectByChatIDGetter projectByChatIDGetter) option {
	return func(s *Service) {
		s.projectByChatIDGetter = projectByChatIDGetter
	}
}

func WithProjectsGetter(projectsGetter projectsGetter) option {
	return func(s *Service) {
		s.projectsGetter = projectsGetter
	}
}

func WithTaskCreator(taskCreator taskCreator) option {
	return func(s *Service) {
		s.taskCreator = taskCreator
	}
}

func WithChatToProjectLinker(chatToProjectLinker chatToProjectLinker) option {
	return func(s *Service) {
		s.chatToProjectLinker = chatToProjectLinker
	}
}

func WithFeedbackLister(feedbackLister feedbackLister) option {
	return func(s *Service) {
		s.feedbackLister = feedbackLister
	}
}

func WithBaseService(baseService baseService) option {
	return func(s *Service) {
		s.taskCreator = baseService
		s.projectsGetter = baseService
		s.feedbackLister = baseService
	}
}

func WithMessageMetaSetter(messageMetaSetter messageMetaSetter) option {
	return func(s *Service) {
		s.messageMetaSetter = messageMetaSetter
	}
}

func WithMessageMetaScanner(messageMetaScanner messageMetaScanner) option {
	return func(s *Service) {
		s.messageMetaScanner = messageMetaScanner
	}
}

func WithRepository(repository repository) option {
	return func(s *Service) {
		s.chatToProjectLinker = repository
		s.projectByChatIDGetter = repository
		s.messageMetaSetter = repository
		s.messageMetaScanner = repository
		s.tgMessageSaver = repository
		s.employeeTgIDUpdater = repository
		s.employeeTgIDByIDGetter = repository
		s.repository = repository
	}
}

func WithFeedbackAnswerer(feedbackAnswerer feedbackAnswerer) option {
	return func(s *Service) {
		s.feedbackAnswerer = feedbackAnswerer
	}
}

func WithFeedbackCreator(feedbackCreator feedbackCreator) option {
	return func(s *Service) {
		s.feedbackCreator = feedbackCreator
	}
}

func WithNotionRepo(notionRepo notionRepo) option {
	return func(s *Service) {
		s.feedbackAnswerer = notionRepo
		s.feedbackCreator = notionRepo
		s.topicsGetter = notionRepo
		s.notionRepo = notionRepo
	}
}

func WithTgRepo(tgRepo tgMessageSender) option {
	return func(s *Service) {
		s.tgRepo = tgRepo
	}
}
