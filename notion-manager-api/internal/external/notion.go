package external

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"math"
	"os"
	"strings"
	"time"

	"github.com/Corray333/employee_dashboard/internal/entities"
	"github.com/Corray333/employee_dashboard/pkg/mindmap"
	"github.com/Corray333/employee_dashboard/pkg/notion"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/spf13/viper"
)

type External struct {
	tg *TelegramClient
}

type TelegramClient struct {
	bot *tgbotapi.BotAPI
}

func (t *TelegramClient) GetBot() *tgbotapi.BotAPI {
	return t.bot
}

func NewClient(token string) *TelegramClient {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal("failed to create bot: ", err)
	}

	bot.Debug = true

	return &TelegramClient{
		bot: bot,
	}
}

func New() *External {
	return &External{
		tg: NewClient(os.Getenv("BOT_TOKEN")),
	}
}

// Employee
type Employee struct {
	ID             string `json:"id"`
	LastEditedTime string `json:"last_edited_time"`
	Icon           struct {
		Type     string   `json:"type"`
		External external `json:"external"`
		File     file     `json:"file"`
		Emoji    emoji    `json:"emoji"`
	} `json:"icon"`
	Properties struct {
		NotificationFlags struct {
			MultiSelect []struct {
				Name string `json:"name"`
			} `json:"multi_select"`
		} `json:"NF"`
		Link struct {
			ID     string `json:"id"`
			Type   string `json:"type"`
			People []struct {
				Object    string `json:"object"`
				ID        string `json:"id"`
				Name      string `json:"name"`
				AvatarURL string `json:"avatar_url"`
				Type      string `json:"type"`
				Person    struct {
					Email string `json:"email"`
				} `json:"person"`
			} `json:"people"`
		} `json:"Ссылка"`
		Salary struct {
			ID     string `json:"id"`
			Type   string `json:"type"`
			Number int    `json:"number"`
		} `json:"Ставка в час"`
		Direction struct {
			Rollup struct {
				Array []struct {
					Select struct {
						Name string `json:"name"`
					} `json:"select"`
				} `json:"array"`
			} `json:"rollup"`
		} `json:"Направление"`
		Name struct {
			Title []struct {
				PlainText string `json:"plain_text"`
			} `json:"title"`
		} `json:"Имя"`
		FIO struct {
			RichText []struct {
				PlainText string `json:"plain_text"`
			} `json:"rich_text"`
		} `json:"ФИО"`
		Telegram struct {
			RichText []struct {
				PlainText string `json:"plain_text"`
			} `json:"rich_text"`
		} `json:"Telegram"`
		Location struct {
			MultiSelect []struct {
				Name string `json:"name"`
			} `json:"multi_select"`
		} `json:"Местоположение"`
		Expertise struct {
			ID       string `json:"id"`
			Type     string `json:"type"`
			Relation []struct {
				ID string `json:"id"`
			} `json:"relation"`
			HasMore bool `json:"has_more"`
		} `json:"Экспертиза"`
		Status struct {
			ID     string `json:"id"`
			Type   string `json:"type"`
			Status struct {
				ID    string `json:"id"`
				Name  string `json:"name"`
				Color string `json:"color"`
			} `json:"status"`
		} `json:"Статус"`
		PhoneNumber struct {
			ID    string `json:"id"`
			Type  string `json:"type"`
			Phone string `json:"phone_number"`
		} `json:"Номер телефона"`
	}
}

func (e *External) GetSheetsTimes(lastSynced int64, projectID string, cursor string) ([]Time, error) {
	req := map[string]interface{}{}

	// TODO: refactor
	if projectID != "" {
		req = map[string]interface{}{
			"filter": map[string]interface{}{
				"and": []map[string]interface{}{
					{
						"timestamp": "last_edited_time",
						"last_edited_time": map[string]interface{}{
							"after": time.Unix(lastSynced, 0).Format(notion.TIME_LAYOUT_IN),
						},
					},
					{
						"property": "Проект",
						"rollup": map[string]interface{}{
							"any": map[string]interface{}{
								"relation": map[string]interface{}{
									"contains": projectID,
								},
							},
						},
					},
				},
			},
			"sorts": []map[string]interface{}{
				{
					"timestamp": "created_time",
					"direction": "ascending",
				},
			},
		}
	} else {
		req = map[string]interface{}{
			"filter": map[string]interface{}{
				"timestamp": "last_edited_time",
				"last_edited_time": map[string]interface{}{
					"after": time.Unix(lastSynced, 0).Format(notion.TIME_LAYOUT),
				},
			},
			"sorts": []map[string]interface{}{
				{
					"timestamp": "created_time",
					"direction": "ascending",
				},
			},
		}
	}

	if cursor != "" {
		fmt.Println("Next cursor applied")
		req["start_cursor"] = cursor
	}

	resp, err := notion.SearchPages(viper.GetString("notion.databases.times"), req)
	if err != nil {
		return nil, err
	}
	times := struct {
		Results    []Time `json:"results"`
		HasMore    bool   `json:"has_more"`
		NextCursor string `json:"next_cursor"`
	}{}

	err = json.Unmarshal(resp, &times)
	if err != nil {
		return nil, err
	}

	if times.HasMore {
		moreTimes, err := e.GetSheetsTimes(lastSynced, projectID, times.NextCursor)
		if err != nil {
			return nil, err
		}
		return append(times.Results, moreTimes...), nil
	}

	return times.Results, nil
}

func (e *External) GetEmployees(lastSynced int64) (employees []entities.Employee, lastUpdate int64, err error) {
	filter := map[string]interface{}{
		"filter": map[string]interface{}{
			"timestamp": "last_edited_time",
			"last_edited_time": map[string]interface{}{
				"after": time.Unix(lastSynced, 0).Format(notion.TIME_LAYOUT),
			},
		},
		"sorts": []map[string]interface{}{
			{
				"timestamp": "created_time",
				"direction": "ascending",
			},
		},
	}

	resp, err := notion.SearchPages(viper.GetString("notion.databases.employees"), filter)
	if err != nil {
		return nil, 0, err
	}
	worker := struct {
		Results []Employee `json:"results"`
	}{}

	if err := json.Unmarshal(resp, &worker); err != nil {
		return nil, 0, err
	}

	lastUpdate = lastSynced

	employees = []entities.Employee{}
	for _, w := range worker.Results {
		employees = append(employees, entities.Employee{
			ID: func() string {
				if len(w.Properties.Link.People) == 0 {
					return ""
				} else {
					return w.Properties.Link.People[0].ID
				}
			}(),
			Username: func() string {
				if len(w.Properties.Name.Title) == 0 {
					return ""
				}
				return w.Properties.Name.Title[0].PlainText
			}(),
			Icon: func() string {
				if len(w.Properties.Link.People) > 0 {
					return w.Properties.Link.People[0].AvatarURL
				}
				return ""
			}(),
			Email: func() string {
				if len(w.Properties.Link.People) == 0 {
					return ""
				}
				return w.Properties.Link.People[0].Person.Email
			}(),
			ProfileID: w.ID,
			NotificationFlags: func() []string {
				flags := []string{}
				for _, flag := range w.Properties.NotificationFlags.MultiSelect {
					flags = append(flags, flag.Name)
				}
				return flags
			}(),
			FIO: func() string {
				if len(w.Properties.FIO.RichText) == 0 {
					return ""
				}
				result := ""
				for _, text := range w.Properties.FIO.RichText {
					result += text.PlainText
				}
				return result
			}(),
			Telegram: func() string {
				if len(w.Properties.Telegram.RichText) == 0 {
					return ""
				}
				result := ""
				for _, text := range w.Properties.Telegram.RichText {
					result += text.PlainText
				}
				return strings.TrimPrefix(result, "@")
			}(),
			Geo: func() string {
				if len(w.Properties.Location.MultiSelect) == 0 {
					return ""
				}
				locations := []string{}
				for _, loc := range w.Properties.Location.MultiSelect {
					locations = append(locations, loc.Name)
				}
				return strings.Join(locations, ",")
			}(),
			ExpertiseID: func() string {
				var expertiseIDs []string
				for _, relation := range w.Properties.Expertise.Relation {
					expertiseIDs = append(expertiseIDs, relation.ID)
				}
				return strings.Join(expertiseIDs, ", ")
			}(),
			Direction: func() string {
				if len(w.Properties.Direction.Rollup.Array) == 0 {
					return ""
				}
				return w.Properties.Direction.Rollup.Array[0].Select.Name
			}(),
			Status: w.Properties.Status.Status.Name,
			Phone:  w.Properties.PhoneNumber.Phone,
		})

		lastEditedTime, err := time.Parse(notion.TIME_LAYOUT_IN, w.LastEditedTime)
		if err != nil {
			return nil, 0, err
		}

		lastUpdate = lastEditedTime.Unix()
	}
	return employees, lastUpdate, nil
}

type Task struct {
	ID             string `json:"id"`
	LastEditedTime string `json:"last_edited_time"`
	Properties     struct {
		Tags struct {
			MultiSelect []struct {
				Name string `json:"name"`
			} `json:"multi_select"`
		} `json:"Теги"`
		Status struct {
			Status struct {
				Name string `json:"name"`
			} `json:"status"`
		} `json:"Статус"`
		ParentTask struct {
			Relation []struct {
				ID string `json:"id"`
			} `json:"relation"`
		} `json:"Родительская задача"`
		Priority struct {
			Select struct {
				Name string `json:"name"`
			} `json:"select"`
		} `json:"Приоритет"`
		Worker struct {
			People []struct {
				Name string `json:"name"`
				ID   string `json:"id"`
			} `json:"people"`
		} `json:"Исполнитель"`
		Product struct {
			Relation []struct {
				ID string `json:"id"`
			} `json:"relation"`
		} `json:"Продукт"`
		Estimated struct {
			Number float64 `json:"number"`
		} `json:"Оценка"`
		Subtasks struct {
			Relation []struct {
				ID string `json:"id"`
			} `json:"relation"`
		} `json:"Подзадачи"`
		Deadline struct {
			Date struct {
				Start *string `json:"start"`
				End   *string `json:"end"`
			} `json:"date"`
		} `json:"Дедлайн"`
		Task struct {
			Title []struct {
				PlainText string `json:"plain_text"`
			} `json:"title"`
		} `json:"Task"`
	} `json:"properties"`
}

func parseDate(dateStr string) (time.Time, error) {
	// Список возможных форматов
	formats := []string{
		time.RFC3339, // "2025-02-23T00:00:00.000+03:00"
		"2006-01-02", // "2025-02-23"
	}

	var parsedTime time.Time
	var err error

	// Пробуем каждый формат
	for _, format := range formats {
		parsedTime, err = time.Parse(format, dateStr)
		if err == nil {
			return parsedTime, nil
		}
	}

	return time.Time{}, fmt.Errorf("не удалось разобрать дату: %s", dateStr)
}

func (e *External) GetTasks(timeFilterType string, lastSynced int64, startCursor string, useTitleFilter bool) (tasks []entities.Task, lastUpdate int64, err error) {
	filter := buildFilter(timeFilterType, lastSynced, startCursor, useTitleFilter)

	lastUpdate = 0
	resp, err := notion.SearchPages(viper.GetString("notion.databases.tasks"), filter)
	if err != nil {
		return nil, 0, err
	}
	task := struct {
		Results    []Task `json:"results"`
		HasMore    bool   `json:"has_more"`
		NextCursor string `json:"next_cursor"`
	}{}

	json.Unmarshal(resp, &task)

	tasks = []entities.Task{}
	for _, w := range task.Results {

		tasks = append(tasks, entities.Task{
			ID: w.ID,
			Employee: func() string {
				if len(w.Properties.Worker.People) == 0 {
					return ""
				}
				return w.Properties.Worker.People[0].Name
			}(),
			Title: func() string {
				if len(w.Properties.Task.Title) == 0 {
					return ""
				}
				title := ""
				for _, t := range w.Properties.Task.Title {
					title += t.PlainText
				}

				return title
			}(),
			Tags: func() []string {
				tags := []string{}
				for _, tag := range w.Properties.Tags.MultiSelect {
					tags = append(tags, tag.Name)
				}
				return tags
			}(),
			Status: w.Properties.Status.Status.Name,
			ProjectID: func() string {
				if len(w.Properties.Product.Relation) == 0 {
					return ""
				}
				return w.Properties.Product.Relation[0].ID
			}(),
			EmployeeID: func() string {
				if len(w.Properties.Worker.People) == 0 {
					return ""
				}
				return w.Properties.Worker.People[0].ID
			}(),
			StartTime: func() int64 {
				if w.Properties.Deadline.Date.Start == nil {
					return 0
				}

				startTime, err := parseDate(*w.Properties.Deadline.Date.Start)
				if err != nil {
					return 0
				}
				return startTime.Unix()
			}(),
			EndTime: func() int64 {
				if w.Properties.Deadline.Date.End == nil {
					return 0
				}

				endTime, err := parseDate(*w.Properties.Deadline.Date.End)
				if err != nil {
					return 0
				}
				return endTime.Unix()
			}(),
		})

		lastEditedTime, err := time.Parse(notion.TIME_LAYOUT_IN, w.LastEditedTime)
		if err != nil {
			return nil, 0, err
		}

		lastUpdate = lastEditedTime.Unix()
	}

	if task.HasMore {
		fmt.Println("has more")
		nextTasks, lastEditedTime, err := e.GetTasks(timeFilterType, lastSynced, task.NextCursor, useTitleFilter)
		if err != nil {
			return nil, 0, err
		}
		lastUpdate = lastEditedTime
		tasks = append(tasks, nextTasks...)
	}

	return tasks, lastUpdate, nil
}

func (e *External) GetNotCorrectPersonTimes() (times []entities.Time, lastUpdate int64, err error) {
	filter := map[string]interface{}{
		"filter": map[string]interface{}{
			"property": "PC-B",
			"formula": map[string]interface{}{
				"checkbox": map[string]interface{}{
					"equals": false,
				},
			},
		},
		"sorts": []map[string]interface{}{
			{
				"timestamp": "last_edited_time",
				"direction": "ascending",
			},
		},
	}

	resp, err := notion.SearchPages(viper.GetString("notion.databases.times"), filter)
	if err != nil {
		return nil, 0, err
	}
	timeResults := struct {
		Results    []Time `json:"results"`
		HasMore    bool   `json:"has_more"`
		NextCursor string `json:"next_cursor"`
	}{}
	if err := json.Unmarshal(resp, &timeResults); err != nil {
		return nil, 0, err
	}

	times = []entities.Time{}
	for _, w := range timeResults.Results {
		times = append(times, entities.Time{
			Description: func() string {
				if len(w.Properties.WhatDid.Title) == 0 {
					return ""
				}

				return w.Properties.WhatDid.Title[0].PlainText
			}(),
			ID: strings.ReplaceAll(w.ID, "-", ""),
			Employee: func() string {
				if len(w.Properties.WhoDid.People) == 0 {
					return ""
				}
				return w.Properties.WhoDid.People[0].Name
			}(),
			EmployeeID: func() string {
				if len(w.Properties.WhoDid.People) == 0 {
					return ""
				}
				return w.Properties.WhoDid.People[0].ID
			}(),
		})

		lastEditedTime, err := time.Parse(notion.TIME_LAYOUT_IN, w.LastEditedTime)
		if err != nil {
			return nil, 0, err
		}

		lastUpdate = lastEditedTime.Unix()
	}

	return times, lastUpdate, nil
}

func (e *External) SetProfileInTime(timeID, profileID string) error {
	req := map[string]interface{}{
		"Person": map[string]interface{}{
			"relation": []map[string]interface{}{
				{
					"id": profileID,
				},
			},
		},
	}

	_, err := notion.UpdatePage(timeID, req)
	if err != nil {
		slog.Error("error updating time page in notion: " + err.Error())
		return err
	}
	return nil
}

func buildFilter(timeFilterType string, lastSynced int64, startCursor string, useTitleFilter bool) map[string]interface{} {
	filter := map[string]interface{}{
		"filter": map[string]interface{}{
			"timestamp": timeFilterType,
			timeFilterType: map[string]interface{}{
				"after": time.Unix(lastSynced, 0).Format(notion.TIME_LAYOUT),
			},
		},
		"sorts": []map[string]interface{}{
			{
				"timestamp": "last_edited_time",
				"direction": "ascending",
			},
		},
	}

	if useTitleFilter {
		forbiddenWords := []string{
			"фикс", "пофиксить", "фиксить", "правка", "править", "поправить", "исправить", "правки", "исправление", "баг", "безуспешно", "разобраться",
		}

		titleFilter := []map[string]interface{}{}
		for _, word := range forbiddenWords {
			titleFilter = append(titleFilter, map[string]interface{}{
				"property": "Task",
				"rich_text": map[string]interface{}{
					"contains": word,
				},
			})
		}

		filter["filter"] = map[string]interface{}{
			"and": []map[string]interface{}{
				filter["filter"].(map[string]interface{}),
				{
					"or": titleFilter,
				},
			},
		}
	}

	if startCursor != "" {
		filter["start_cursor"] = startCursor
	}

	return filter
}

type Project struct {
	ID             string `json:"id"`
	CreatedTime    string `json:"created_time"`
	LastEditedTime string `json:"last_edited_time"`
	Icon           struct {
		Type     string   `json:"type"`
		External external `json:"external"`
		File     file     `json:"file"`
		Emoji    string   `json:"emoji"`
	} `json:"icon"`
	Properties struct {
		Name struct {
			Title []struct {
				PlainText string `json:"plain_text"`
			} `json:"title"`
		} `json:"Name"`
		Status struct {
			Status struct {
				Name  string `json:"name"`
				ID    string `json:"id"`
				Color string `json:"color"`
			} `json:"status"`
		} `json:"Статус"`
		ProjectType struct {
			ID     string `json:"id"`
			Type   string `json:"type"`
			Select struct {
				Name string `json:"name"`
			} `json:"select"`
		} `json:"Тип проекта"`
		Manager struct {
			ID       string `json:"id"`
			Type     string `json:"type"`
			Relation []struct {
				ID string `json:"id"`
			} `json:"relation"`
			HasMore bool `json:"has_more"`
		} `json:"Менеджер"`
		ManagerLink struct {
			ID     string `json:"id"`
			Type   string `json:"type"`
			Rollup struct {
				Type     string        `json:"type"`
				Array    []interface{} `json:"array"` // Определите структуру при необходимости
				Function string        `json:"function"`
			} `json:"rollup"`
		} `json:"Менеджер Link"`
	} `json:"properties"`
}

type external struct {
	Url string `json:"url"`
}

type file struct {
	Url string `json:"url"`
}

type emoji struct {
	Emoji string `json:"emoji"`
}

func (e *External) GetProjects(lastSynced int64) (projects []entities.Project, lastUpdate int64, err error) {
	filter := map[string]interface{}{
		"filter": map[string]interface{}{
			"timestamp": "last_edited_time",
			"last_edited_time": map[string]interface{}{
				"after": time.Unix(lastSynced, 0).Format(notion.TIME_LAYOUT),
			},
		},
		"sorts": []map[string]interface{}{
			{
				"timestamp": "created_time",
				"direction": "ascending",
			},
		},
	}

	resp, err := notion.SearchPages(viper.GetString("notion.databases.projects"), filter)
	if err != nil {
		return nil, 0, err
	}
	project := struct {
		Results []Project `json:"results"`
	}{}
	if err := json.Unmarshal(resp, &project); err != nil {
		slog.Error("Error unmarshalling projects", slog.String("error", err.Error()))
		return nil, 0, err
	}

	projects = []entities.Project{}
	for _, w := range project.Results {
		projects = append(projects, entities.Project{
			ID: w.ID,
			Name: func() string {
				if len(w.Properties.Name.Title) == 0 {
					return ""
				}
				return w.Properties.Name.Title[0].PlainText
			}(),
			Icon: func() string {
				if w.Icon.Type == "emoji" {
					return w.Icon.Emoji
				}
				if w.Icon.Type == "external" {
					return w.Icon.External.Url
				}
				if w.Icon.Type == "file" {
					return w.Icon.File.Url
				}
				return ""
			}(),
			IconType: w.Icon.Type,
			Status:   w.Properties.Status.Status.Name,
			ManagerID: func() string {
				if len(w.Properties.Manager.Relation) == 0 {
					return ""
				}
				return w.Properties.Manager.Relation[0].ID
			}(),
			Type: w.Properties.ProjectType.Select.Name,
		})

		lastEditedTime, err := time.Parse(notion.TIME_LAYOUT_IN, w.LastEditedTime)
		if err != nil {
			return nil, 0, err
		}

		lastUpdate = lastEditedTime.Unix()
	}
	return projects, lastUpdate, nil
}

func (e *External) WriteOfTime(timeToWriteOf *entities.TimeMsg) error {

	req := map[string]interface{}{
		"Всего ч": map[string]interface{}{
			"number": math.Ceil((float64(timeToWriteOf.Duration)/60/60)/0.15) * 0.15,
		},
		"Задача": map[string]interface{}{
			"relation": []map[string]interface{}{
				{
					"id": timeToWriteOf.TaskID,
				},
			},
		},
		"Что делали": map[string]interface{}{
			"type": "title",
			"title": []map[string]interface{}{
				{
					"type": "text",
					"text": map[string]interface{}{
						"content": timeToWriteOf.Description,
					},
				},
			},
		},
		"Дата работ": map[string]interface{}{
			"type": "date",
			"date": map[string]interface{}{
				"start": time.Now().Format(notion.TIME_LAYOUT),
			},
		},
		"Исполнитель": map[string]interface{}{
			"people": []map[string]interface{}{
				{
					"object": "user",
					"id":     timeToWriteOf.EmployeeID,
				},
			},
		},
	}

	_, err := notion.CreatePage(viper.GetString("notion.databases.times"), req, nil, "")
	if err != nil {
		slog.Error("error creating time page in notion: " + err.Error())
		return err
	}
	return nil
}

type Time struct {
	ID             string `json:"id"`
	CreatedTime    string `json:"created_time"`
	LastEditedTime string `json:"last_edited_time"`
	Properties     struct {
		TotalHours struct {
			Number float64 `json:"number"`
		} `json:"Всего ч"`
		Analytics struct {
			Relation []struct{} `json:"relation"`
		} `json:"Аналитика"`
		PayableHours struct {
			Formula struct {
				Number float64 `json:"number"`
			} `json:"formula"`
		} `json:"К оплате ч."`
		Task struct {
			Relation []struct {
				ID string `json:"id"`
			} `json:"relation"`
		} `json:"Задача"`
		Direction struct {
			Select struct {
				Name string `json:"name"`
			} `json:"select"`
		} `json:"Направление"`
		TaskName struct {
			Formula struct {
				String string `json:"string"`
			} `json:"formula"`
		} `json:"Название задачи"`
		WorkDate struct {
			Date struct {
				Start    string      `json:"start"`
				End      interface{} `json:"end"`
				TimeZone interface{} `json:"time_zone"`
			} `json:"date"`
		} `json:"Дата работ"`
		WhoDid struct {
			People []struct {
				Name string `json:"name"`
				ID   string `json:"id"`
			} `json:"people"`
		} `json:"Исполнитель"`
		EstimateHours struct {
			Formula struct {
				String string `json:"string"`
			} `json:"formula"`
		} `json:"Оценка ч"`
		CreatedTimeField struct {
			CreatedTime string `json:"created_time"`
		} `json:"Created time"`
		Payment struct {
			Checkbox bool `json:"checkbox"`
		} `json:"Оплата"`
		Project struct {
			Rollup struct {
				Array []struct {
					Relation []struct {
						ID string `json:"id"`
					} `json:"relation"`
				} `json:"array"`
			} `json:"rollup"`
		} `json:"Проект"`
		StatusHours struct {
			Formula struct {
				String string `json:"string"`
			} `json:"formula"`
		} `json:"Статус ч"`
		Month struct {
			Formula struct {
				String string `json:"string"`
			} `json:"formula"`
		} `json:"Месяц"`
		ProjectName struct {
			Formula struct {
				String string `json:"string"`
			} `json:"formula"`
		} `json:"Имя проекта"`
		ProjectStatus struct {
			Formula struct {
				String string `json:"string"`
			} `json:"formula"`
		} `json:"Статус проекта"`
		WhatDid struct {
			Title []struct {
				PlainText string `json:"plain_text"`
			} `json:"title"`
		} `json:"Что делали"`
		BH struct {
			Formula struct {
				Number float64 `json:"number"`
			} `json:"formula"`
		} `json:"BH"`
		SH struct {
			Number float64 `json:"number"` // Number or null
		} `json:"SH"`
		DH struct {
			Number float64 `json:"number"` // Number or null
		} `json:"DH"`
		BHGS struct {
			Formula struct {
				String string `json:"string"`
			} `json:"formula"`
		} `json:"BHGS"`
		WeekNumber struct {
			Formula struct {
				Number float64 `json:"number"`
			} `json:"formula"`
		} `json:"Номер недели"`
		DayNumber struct {
			Formula struct {
				Number float64 `json:"number"`
			} `json:"formula"`
		} `json:"Номер дня"`
		MonthNumber struct {
			Formula struct {
				Number float64 `json:"number"`
			} `json:"formula"`
		} `json:"Номер месяца"`
		// Добавленные поля:
		PH struct {
			Formula struct {
				Number float64 `json:"number"`
			} `json:"formula"`
		} `json:"PH"`
		Expertise struct {
			Rollup struct {
				Array []struct {
					Relation []struct {
						ID string `json:"id"`
					} `json:"relation"`
				} `json:"array"`
			} `json:"rollup"`
		} `json:"Экспертиза"`
		Overtime struct {
			Checkbox bool `json:"checkbox"`
		} `json:"Сверхурочные"`
		PCB struct {
			Formula struct {
				Boolean bool `json:"boolean"`
			} `json:"formula"`
		} `json:"PC-B"`
		TaskEstimate struct {
			Formula struct {
				String string `json:"string"`
			} `json:"formula"`
		} `json:"Оценка задачи"`
		Person struct {
			Relation []struct {
				ID string `json:"id"`
			} `json:"relation"`
		} `json:"Person"`
		IDField struct {
			UniqueID struct {
				Prefix string `json:"prefix"`
				Number int    `json:"number"`
			} `json:"unique_id"`
		} `json:"ID"`
		ET struct {
			Formula struct {
				String string `json:"string"`
			} `json:"formula"`
		} `json:"ET"`
		Priority struct {
			Rollup struct {
				Array []struct {
					Select struct {
						Name string `json:"name"`
					} `json:"select"`
				} `json:"array"`
			} `json:"rollup"`
		} `json:"Приоритет"`
		MainTask struct {
			Rollup struct {
				Array []struct {
					Formula struct {
						String string `json:"string"`
					} `json:"formula"`
				} `json:"array"`
			} `json:"rollup"`
		} `json:"Главная задача"`
		TargetTask struct {
			Formula struct {
				String string `json:"string"`
			} `json:"formula"`
		} `json:"Целевая задача"`
		CR struct {
			Formula struct {
				Boolean bool `json:"boolean"`
			} `json:"formula"`
		} `json:"CR"`
	} `json:"properties"`
	URL string `json:"url"`
}

func (e *External) GetTimes(timeFilterType string, lastSynced int64, startCursor string, useWhatDidFilter bool) (times []entities.Time, lastUpdate int64, err error) {
	filter := buildTimeFilter(timeFilterType, lastSynced, startCursor, useWhatDidFilter)

	lastUpdate = 0

	resp, err := notion.SearchPages(viper.GetString("notion.databases.times"), filter)
	if err != nil {
		return nil, 0, err
	}
	timeResults := struct {
		Results    []Time `json:"results"`
		HasMore    bool   `json:"has_more"`
		NextCursor string `json:"next_cursor"`
	}{}
	if err := json.Unmarshal(resp, &timeResults); err != nil {
		return nil, 0, err
	}

	times = []entities.Time{}
	for _, w := range timeResults.Results {
		times = append(times, entities.Time{
			Description: func() string {
				if len(w.Properties.WhatDid.Title) == 0 {
					return ""
				}

				return w.Properties.WhatDid.Title[0].PlainText
			}(),
			ID: strings.ReplaceAll(w.ID, "-", ""),
			Employee: func() string {
				if len(w.Properties.WhoDid.People) == 0 {
					return ""
				}
				return w.Properties.WhoDid.People[0].Name
			}(),
			EmployeeID: func() string {
				if len(w.Properties.WhoDid.People) == 0 {
					return ""
				}
				return w.Properties.WhoDid.People[0].ID
			}(),
		})

		lastEditedTime, err := time.Parse(notion.TIME_LAYOUT_IN, w.LastEditedTime)
		if err != nil {
			return nil, 0, err
		}

		lastUpdate = lastEditedTime.Unix()
	}

	if timeResults.HasMore {
		fmt.Println("time has more")
		nextTasks, lastEditedTime, err := e.GetTimes(timeFilterType, lastSynced, timeResults.NextCursor, useWhatDidFilter)
		if err != nil {
			return nil, 0, err
		}
		lastUpdate = lastEditedTime
		times = append(times, nextTasks...)
	}

	return times, lastUpdate, nil
}

func buildTimeFilter(timeFilterType string, lastSynced int64, startCursor string, useWhatDidFilter bool) map[string]interface{} {
	filter := map[string]interface{}{
		"filter": map[string]interface{}{
			"timestamp": timeFilterType,
			timeFilterType: map[string]interface{}{
				"after": time.Unix(lastSynced, 0).Format(notion.TIME_LAYOUT),
			},
		},
		"sorts": []map[string]interface{}{
			{
				"timestamp": "last_edited_time",
				"direction": "ascending",
			},
		},
	}

	if useWhatDidFilter {
		forbiddenWords := []string{
			"фикс", "пофиксить", "фиксить", "правка", "править", "поправить", "исправить", "правки", "исправление", "баг", "безуспешно", "разобраться",
		}

		whatDidFilter := []map[string]interface{}{}
		for _, word := range forbiddenWords {
			whatDidFilter = append(whatDidFilter, map[string]interface{}{
				"property": "Что делали",
				"rich_text": map[string]interface{}{
					"contains": word,
				},
			})
		}

		filter["filter"] = map[string]interface{}{
			"and": []map[string]interface{}{
				filter["filter"].(map[string]interface{}),
				{
					"or": whatDidFilter,
				},
			},
		}
	}

	if startCursor != "" {
		filter["start_cursor"] = startCursor
	}

	return filter
}

type Projects struct {
	Results []struct {
		ID string `json:"id"`
	}
}

func (e *External) CreateMindmapTasks(projectName string, tasks []mindmap.Task) error {
	fmt.Println("Creating tasks for project:", projectName)

	fmt.Println(projectName)
	// Поиск проекта в Notion
	projectFilter := map[string]interface{}{
		"filter": map[string]interface{}{
			"property": "Name",
			"rich_text": map[string]interface{}{
				"contains": projectName,
			},
		},
	}
	projectsResp, err := notion.SearchPages(os.Getenv("PROJECTS_DB"), projectFilter)
	if err != nil {
		slog.Error("Notion error while searching projects: " + err.Error())
		return err
	}

	// Извлекаем ID проекта
	projects := Projects{}
	if err := json.Unmarshal(projectsResp, &projects); err != nil {
		slog.Error("Error unmarshalling projects response: " + err.Error())
		return err
	}

	projectID := ""
	if len(projects.Results) != 0 {
		projectID = projects.Results[0].ID
	}

	// Создаем задачи для проекта
	for _, task := range tasks {
		if err := createMindmapTask(projectID, &task, "", 0); err != nil {
			return err
		}
	}

	return nil
}

type PageCreated struct {
	ID string `json:"id"`
}

func createMindmapTask(projectID string, task *mindmap.Task, parentID string, level int) error {
	fmt.Printf("Creating task: %s\n", task.Title)

	// Основная структура для страницы задачи
	req := map[string]interface{}{
		"Оценка": map[string]interface{}{
			"number": task.Hours,
		},
		"Task": map[string]interface{}{
			"type": "title",
			"title": []map[string]interface{}{
				{
					"type": "text",
					"text": map[string]interface{}{
						"content": task.Title,
					},
				},
			},
		},
	}

	// Если проект указан, связываем задачу с проектом
	if projectID != "" {
		req["Продукт"] = map[string]interface{}{
			"relation": []map[string]interface{}{
				{
					"id": projectID,
				},
			},
		}
	}

	content := []map[string]interface{}{}
	// Если задача второго уровня, добавляем родительскую задачу
	if parentID != "" {
		req["Родительская задача"] = map[string]interface{}{
			"relation": []map[string]interface{}{
				{
					"id": parentID,
				},
			},
		}

		// Массив для вложенных задач, если они есть
		content = createCheckboxes(task.Subtasks)
	}

	// Создаем страницу задачи в Notion
	resp, err := notion.CreatePage(viper.GetString("notion.databases.tasks"), req, content, "")
	if err != nil {
		slog.Error("Notion error while creating task: " + err.Error())
		return err
	}

	var page PageCreated
	if err := json.Unmarshal(resp, &page); err != nil {
		slog.Error("Error unmarshalling response: " + err.Error())
		return err
	}

	if level == 0 {
		// Рекурсивно создаем подзадачи
		for _, subtask := range task.Subtasks {
			if err := createMindmapTask(projectID, &subtask, page.ID, level+1); err != nil {
				return err
			}
		}
	}

	return nil
}

func createCheckboxes(tasks []mindmap.Task) []map[string]interface{} {
	content := []map[string]interface{}{}

	for _, subtask := range tasks {
		content = append(content, map[string]interface{}{
			"type": "to_do",
			"to_do": map[string]interface{}{
				"rich_text": []map[string]interface{}{
					{
						"type": "text",
						"text": map[string]interface{}{
							"content": subtask.Title,
						},
					},
				},
				"checked":  false,
				"children": createCheckboxes(subtask.Subtasks),
			},
		})
	}

	return content
}

type Expertise struct {
	ID         string `json:"id"`
	Properties struct {
		Name struct {
			Title []struct {
				PlainText string `json:"plain_text"`
			} `json:"title"`
		} `json:"Name"`
		Tag struct {
			Select struct {
				Name string `json:"name"`
			} `json:"select"`
		} `json:"Tag"`
		Description struct {
			RichText []struct {
				PlainText string `json:"plain_text"`
			} `json:"rich_text"`
		} `json:"Описание"`
	} `json:"properties"`
}

func (e *External) GetExpertise() (expertises []entities.Expertise, err error) {
	filter := map[string]interface{}{}

	resp, err := notion.SearchPages(viper.GetString("notion.databases.expertise"), filter)
	if err != nil {
		return nil, err
	}
	expertiseResults := struct {
		Results []Expertise `json:"results"`
	}{}
	if err := json.Unmarshal(resp, &expertiseResults); err != nil {
		slog.Error("Error unmarshalling expertise", slog.String("error", err.Error()))
		return nil, err
	}

	expertises = []entities.Expertise{}
	for _, w := range expertiseResults.Results {
		expertises = append(expertises, entities.Expertise{
			ID: w.ID,
			Name: func() string {
				if len(w.Properties.Name.Title) == 0 {
					return ""
				}
				return w.Properties.Name.Title[0].PlainText
			}(),
			Direction: w.Properties.Tag.Select.Name,
			Description: func() string {
				if len(w.Properties.Description.RichText) == 0 {
					return ""
				}
				return w.Properties.Description.RichText[0].PlainText
			}(),
		})

	}
	return expertises, nil
}
