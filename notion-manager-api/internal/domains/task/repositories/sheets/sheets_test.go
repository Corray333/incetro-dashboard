package sheets

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/Corray333/employee_dashboard/internal/domains/task/entities/task"
	"github.com/google/uuid"
)

func TestGenerateMonthStatusRows(t *testing.T) {
	tests := []struct {
		name     string
		tasks    []task.Task
		expected int // количество ожидаемых строк
		checkRow func([][]interface{}) bool
	}{
		{
			name:  "empty tasks",
			tasks: []task.Task{},
			expected: 0,
		},
		{
			name: "task without start date",
			tasks: []task.Task{
				{
					ID:     uuid.New(),
					Task:   "Test task",
					Status: task.StatusInProgress,
					// Start не установлен (zero time)
				},
			},
			expected: 0,
		},
		{
			name: "single task single month",
			tasks: []task.Task{
				{
					ID:     uuid.New(),
					Task:   "Test task",
					Status: task.StatusInProgress,
					Start:  time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
					End:    time.Date(2024, 1, 25, 0, 0, 0, 0, time.UTC),
				},
			},
			expected: 1,
			checkRow: func(rows [][]interface{}) bool {
				if len(rows) != 1 {
					return false
				}
				row := rows[0]
				// Проверяем статус
				if row[2] != string(task.StatusInProgress) {
					return false
				}
				// Проверяем дату (первый день месяца)
				if row[3] != "01/01/2024" || row[4] != "01/01/2024" {
					return false
				}
				return true
			},
		},
		{
			name: "task spanning multiple months",
			tasks: []task.Task{
				{
					ID:     uuid.New(),
					Task:   "Long task",
					Status: task.StatusInProgress,
					Start:  time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
					End:    time.Date(2024, 3, 10, 0, 0, 0, 0, time.UTC),
				},
			},
			expected: 3, // январь, февраль, март
			checkRow: func(rows [][]interface{}) bool {
				if len(rows) != 3 {
					return false
				}
				// Проверяем, что все строки имеют одинаковый статус
				for _, row := range rows {
					if row[2] != string(task.StatusInProgress) {
						return false
					}
				}
				// Проверяем даты
				expectedDates := []string{"01/01/2024", "01/02/2024", "01/03/2024"}
				actualDates := make([]string, len(rows))
				for i, row := range rows {
					actualDates[i] = row[3].(string)
				}
				// Сортируем для сравнения
				for _, expected := range expectedDates {
					found := false
					for _, actual := range actualDates {
						if expected == actual {
							found = true
							break
						}
					}
					if !found {
						return false
					}
				}
				return true
			},
		},
		{
			name: "multiple tasks same month different status",
			tasks: []task.Task{
				{
					ID:     uuid.New(),
					Task:   "Task 1",
					Status: task.StatusInProgress,
					Start:  time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC),
					End:    time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
				},
				{
					ID:     uuid.New(),
					Task:   "Task 2",
					Status: task.StatusDone,
					Start:  time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC),
					End:    time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC),
				},
			},
			expected: 2, // две разные комбинации месяц-статус
			checkRow: func(rows [][]interface{}) bool {
				if len(rows) != 2 {
					return false
				}
				// Проверяем, что есть обе комбинации статусов
				statuses := make(map[string]bool)
				for _, row := range rows {
					statuses[row[2].(string)] = true
					if row[3] != "01/01/2024" || row[4] != "01/01/2024" {
						return false
					}
				}
				return statuses[string(task.StatusInProgress)] && statuses[string(task.StatusDone)]
			},
		},
		{
			name: "task with zero end date",
			tasks: []task.Task{
				{
					ID:     uuid.New(),
					Task:   "Task with no end",
					Status: task.StatusInProgress,
					Start:  time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
					// End не установлен (zero time)
				},
			},
			expected: 1,
			checkRow: func(rows [][]interface{}) bool {
				if len(rows) != 1 {
					return false
				}
				row := rows[0]
				return row[2] == string(task.StatusInProgress) && row[3] == "01/01/2024"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateMonthStatusRows(tt.tasks)
			if len(result) != tt.expected {
				t.Errorf("generateMonthStatusRows() returned %d rows, expected %d", len(result), tt.expected)
			}
			if tt.checkRow != nil && !tt.checkRow(result) {
				t.Errorf("generateMonthStatusRows() row validation failed")
			}
		})
	}
}

func TestGenerateParentTaskRows(t *testing.T) {
	tests := []struct {
		name     string
		tasks    []task.Task
		expected int
		checkRow func([][]interface{}) bool
	}{
		{
			name:     "empty tasks",
			tasks:    []task.Task{},
			expected: 0,
		},
		{
			name: "task without children",
			tasks: []task.Task{
				{
					ID:         uuid.New(),
					Task:       "Regular task",
					Status:     task.StatusInProgress,
					Start:      time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
					End:        time.Date(2024, 1, 25, 0, 0, 0, 0, time.UTC),
					ChildCount: 0, // нет дочерних задач
				},
			},
			expected: 0,
		},
		{
			name: "parent task without start date",
			tasks: []task.Task{
				{
					ID:         uuid.New(),
					Task:       "Parent task",
					Status:     task.StatusInProgress,
					ChildCount: 3,
					// Start не установлен
				},
			},
			expected: 0,
		},
		{
			name: "parent task single month",
			tasks: []task.Task{
				{
					ID:         uuid.New(),
					Task:       "Parent task",
					Status:     task.StatusInProgress,
					Start:      time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
					End:        time.Date(2024, 1, 25, 0, 0, 0, 0, time.UTC),
					ChildCount: 2,
				},
			},
			expected: 1,
			checkRow: func(rows [][]interface{}) bool {
				if len(rows) != 1 {
					return false
				}
				row := rows[0]
				// Проверяем дату
				if row[3] != "01/01/2024" || row[4] != "01/01/2024" {
					return false
				}
				// Проверяем, что родительская задача содержит гиперссылку
				parentTask := row[5].(string)
				if !strings.Contains(parentTask, "=HYPERLINK") {
					return false
				}
				if !strings.Contains(parentTask, "Parent task") {
					return false
				}
				return true
			},
		},
		{
			name: "parent task spanning multiple months",
			tasks: []task.Task{
				{
					ID:         uuid.New(),
					Task:       "Long parent task",
					Status:     task.StatusInProgress,
					Start:      time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
					End:        time.Date(2024, 3, 10, 0, 0, 0, 0, time.UTC),
					ChildCount: 5,
				},
			},
			expected: 3, // январь, февраль, март
			checkRow: func(rows [][]interface{}) bool {
				if len(rows) != 3 {
					return false
				}
				// Проверяем даты
				expectedDates := []string{"01/01/2024", "01/02/2024", "01/03/2024"}
				actualDates := make([]string, len(rows))
				for i, row := range rows {
					actualDates[i] = row[3].(string)
					// Проверяем гиперссылку в каждой строке
					parentTask := row[5].(string)
					if !strings.Contains(parentTask, "=HYPERLINK") || !strings.Contains(parentTask, "Long parent task") {
						return false
					}
				}
				// Проверяем, что все ожидаемые даты присутствуют
				for _, expected := range expectedDates {
					found := false
					for _, actual := range actualDates {
						if expected == actual {
							found = true
							break
						}
					}
					if !found {
						return false
					}
				}
				return true
			},
		},
		{
			name: "parent task with zero end date",
			tasks: []task.Task{
				{
					ID:         uuid.New(),
					Task:       "Parent with no end",
					Status:     task.StatusInProgress,
					Start:      time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
					ChildCount: 1,
					// End не установлен
				},
			},
			expected: 1,
			checkRow: func(rows [][]interface{}) bool {
				if len(rows) != 1 {
					return false
				}
				row := rows[0]
				return row[3] == "01/01/2024" && strings.Contains(row[5].(string), "Parent with no end")
			},
		},
		{
			name: "multiple parent tasks",
			tasks: []task.Task{
				{
					ID:         uuid.New(),
					Task:       "Parent 1",
					Status:     task.StatusInProgress,
					Start:      time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
					End:        time.Date(2024, 1, 25, 0, 0, 0, 0, time.UTC),
					ChildCount: 2,
				},
				{
					ID:         uuid.New(),
					Task:       "Parent 2",
					Status:     task.StatusDone,
					Start:      time.Date(2024, 2, 5, 0, 0, 0, 0, time.UTC),
					End:        time.Date(2024, 2, 15, 0, 0, 0, 0, time.UTC),
					ChildCount: 1,
				},
				{
					ID:         uuid.New(),
					Task:       "Regular task",
					Status:     task.StatusInProgress,
					Start:      time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC),
					ChildCount: 0, // не родительская
				},
			},
			expected: 2, // только две родительские задачи
			checkRow: func(rows [][]interface{}) bool {
				if len(rows) != 2 {
					return false
				}
				// Проверяем, что обе родительские задачи присутствуют
				parentNames := make(map[string]bool)
				for _, row := range rows {
					parentTask := row[5].(string)
					if strings.Contains(parentTask, "Parent 1") {
						parentNames["Parent 1"] = true
					}
					if strings.Contains(parentTask, "Parent 2") {
						parentNames["Parent 2"] = true
					}
				}
				return parentNames["Parent 1"] && parentNames["Parent 2"]
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateParentTaskRows(tt.tasks)
			if len(result) != tt.expected {
				t.Errorf("generateParentTaskRows() returned %d rows, expected %d", len(result), tt.expected)
			}
			if tt.checkRow != nil && !tt.checkRow(result) {
				t.Errorf("generateParentTaskRows() row validation failed")
			}
		})
	}
}

// TestGenerateMonthStatusRowsStructure проверяет структуру возвращаемых строк
func TestGenerateMonthStatusRowsStructure(t *testing.T) {
	tasks := []task.Task{
		{
			ID:     uuid.New(),
			Task:   "Test task",
			Status: task.StatusInProgress,
			Start:  time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			End:    time.Date(2024, 1, 25, 0, 0, 0, 0, time.UTC),
		},
	}

	result := generateMonthStatusRows(tasks)
	if len(result) != 1 {
		t.Fatalf("Expected 1 row, got %d", len(result))
	}

	row := result[0]
	if len(row) != 11 {
		t.Fatalf("Expected row to have 11 columns, got %d", len(row))
	}

	// Проверяем структуру строки
	expectedStructure := []interface{}{
		"",                              // Название задачи пустое
		"",                              // Приоритет пустой
		string(task.StatusInProgress),   // Статус
		"01/01/2024",                   // startDate
		"01/01/2024",                   // endDate
		"",                              // Родительская задача пустая
		"",                              // Главная задача пустая
		"",                              // Направление пустое
		"",                              // Экспертиза пустая
		"",                              // TotalHours пустое
		"",                              // TotalEstimate пустое
	}

	if !reflect.DeepEqual(row, expectedStructure) {
		t.Errorf("Row structure mismatch.\nExpected: %v\nGot: %v", expectedStructure, row)
	}
}

// TestGenerateParentTaskRowsStructure проверяет структуру возвращаемых строк для родительских задач
func TestGenerateParentTaskRowsStructure(t *testing.T) {
	taskID := uuid.New()
	tasks := []task.Task{
		{
			ID:         taskID,
			Task:       "Parent task",
			Status:     task.StatusInProgress,
			Start:      time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			End:        time.Date(2024, 1, 25, 0, 0, 0, 0, time.UTC),
			ChildCount: 2,
		},
	}

	result := generateParentTaskRows(tasks)
	if len(result) != 1 {
		t.Fatalf("Expected 1 row, got %d", len(result))
	}

	row := result[0]
	if len(row) != 11 {
		t.Fatalf("Expected row to have 11 columns, got %d", len(row))
	}

	// Проверяем основную структуру
	if row[0] != "" || row[1] != "" || row[2] != "" {
		t.Error("First three columns should be empty")
	}

	if row[3] != "01/01/2024" || row[4] != "01/01/2024" {
		t.Error("Start and end dates should be first day of month")
	}

	// Проверяем гиперссылку
	parentTaskLink := row[5].(string)
	expectedURL := "https://notion.so/" + strings.ReplaceAll(taskID.String(), "-", "")
	if !strings.Contains(parentTaskLink, expectedURL) {
		t.Errorf("Parent task link should contain URL %s, got %s", expectedURL, parentTaskLink)
	}

	if !strings.Contains(parentTaskLink, "Parent task") {
		t.Error("Parent task link should contain task name")
	}

	// Проверяем остальные пустые поля
	for i := 6; i < 11; i++ {
		if row[i] != "" {
			t.Errorf("Column %d should be empty, got %v", i, row[i])
		}
	}
}

// TestGenerateMonthStatusRowsEdgeCases проверяет граничные случаи
func TestGenerateMonthStatusRowsEdgeCases(t *testing.T) {
	t.Run("task with quotes in name", func(t *testing.T) {
		// Этот тест проверяет, что функция не падает на задачах с кавычками в названии
		// (хотя название не используется в generateMonthStatusRows)
		tasks := []task.Task{
			{
				ID:     uuid.New(),
				Task:   `Task with "quotes"`,
				Status: task.StatusInProgress,
				Start:  time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			},
		}

		result := generateMonthStatusRows(tasks)
		if len(result) != 1 {
			t.Errorf("Expected 1 row, got %d", len(result))
		}
	})

	t.Run("year boundary crossing", func(t *testing.T) {
		tasks := []task.Task{
			{
				ID:     uuid.New(),
				Task:   "Year crossing task",
				Status: task.StatusInProgress,
				Start:  time.Date(2023, 12, 15, 0, 0, 0, 0, time.UTC),
				End:    time.Date(2024, 2, 10, 0, 0, 0, 0, time.UTC),
			},
		}

		result := generateMonthStatusRows(tasks)
		if len(result) != 3 {
			t.Errorf("Expected 3 rows (Dec 2023, Jan 2024, Feb 2024), got %d", len(result))
		}

		// Проверяем, что есть строки для всех месяцев
		dates := make(map[string]bool)
		for _, row := range result {
			dates[row[3].(string)] = true
		}

		expectedDates := []string{"01/12/2023", "01/01/2024", "01/02/2024"}
		for _, expected := range expectedDates {
			if !dates[expected] {
				t.Errorf("Missing expected date %s", expected)
			}
		}
	})
}

// TestGenerateParentTaskRowsEdgeCases проверяет граничные случаи для родительских задач
func TestGenerateParentTaskRowsEdgeCases(t *testing.T) {
	t.Run("parent task with quotes in name", func(t *testing.T) {
		taskID := uuid.New()
		tasks := []task.Task{
			{
				ID:         taskID,
				Task:       `Parent "quoted" task`,
				Status:     task.StatusInProgress,
				Start:      time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
				ChildCount: 1,
			},
		}

		result := generateParentTaskRows(tasks)
		if len(result) != 1 {
			t.Errorf("Expected 1 row, got %d", len(result))
		}

		// Проверяем, что кавычки правильно экранированы в гиперссылке
		parentTaskLink := result[0][5].(string)
		if !strings.Contains(parentTaskLink, `Parent ""quoted"" task`) {
			t.Errorf("Quotes should be escaped in hyperlink, got: %s", parentTaskLink)
		}
	})

	t.Run("parent task year boundary crossing", func(t *testing.T) {
		tasks := []task.Task{
			{
				ID:         uuid.New(),
				Task:       "Year crossing parent",
				Status:     task.StatusInProgress,
				Start:      time.Date(2023, 11, 15, 0, 0, 0, 0, time.UTC),
				End:        time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC),
				ChildCount: 2,
			},
		}

		result := generateParentTaskRows(tasks)
		if len(result) != 3 {
			t.Errorf("Expected 3 rows (Nov 2023, Dec 2023, Jan 2024), got %d", len(result))
		}

		// Проверяем даты
		dates := make(map[string]bool)
		for _, row := range result {
			dates[row[3].(string)] = true
		}

		expectedDates := []string{"01/11/2023", "01/12/2023", "01/01/2024"}
		for _, expected := range expectedDates {
			if !dates[expected] {
				t.Errorf("Missing expected date %s", expected)
			}
		}
	})
}