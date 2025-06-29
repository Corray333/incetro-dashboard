package main

import (
	"context"
	"fmt"

	"github.com/Corray333/employee_dashboard/internal/app"
	"github.com/Corray333/employee_dashboard/internal/config"
	"github.com/tmc/langchaingo/tools/sqldatabase/postgresql"
)

type SQLDatabaseTool struct {
	db *postgresql.PostgreSQL
}

func (sdt *SQLDatabaseTool) Name() string {
	return "postgresql_database_tool"
}

func (sdt *SQLDatabaseTool) Description() string {
	return "PostgreSQL-only tool. Использует диалект PostgreSQL: TEXT, SERIAL, \\\\d, information_schema. НЕ использовать DESCRIBE. \\dt тоже запрещено, нужно использовать select из служебных таблиц"
}

func (sdt *SQLDatabaseTool) Call(ctx context.Context, input string) (string, error) {
	fmt.Println("Executing SQL query:", input)
	cols, results, err := sdt.db.Query(ctx, input)
	if err != nil {
		return "", err
	}
	result := ""
	for _, row := range results {
		for i, col := range cols {
			result += fmt.Sprintf("%s: %s\n", col, row[i])
		}
		result += "\n"
	}
	return result, nil
}

func main() {

	config.MustInit()
	fmt.Println("Start")

	app.New().Run()

	// // Создание контекста
	// ctx := context.Background()

	// Инициализация языковой модели Google Gemini
	// llm, err := googleai.New(ctx, googleai.WithAPIKey(os.Getenv("GEMINI_API_KEY")), googleai.WithDefaultModel("gemini-2.0-flash"))
	// if err != nil {
	// 	log.Fatalf("Ошибка инициализации LLM: %v", err)
	// }

	// llm, err := ollama.New(
	// 	ollama.WithModel("gemma3n:e4b"),                           // любой slug из `ollama list`
	// 	ollama.WithServerURL("http://host.docker.internal:11434"), // не обязателен, это значение по-умолчанию
	// )
	// if err != nil {
	// 	log.Fatalf("ошибка инициализации LLM: %v", err)
	// }

	// // Получение строки подключения к PostgreSQL из переменной окружения
	// dsn := os.Getenv("LANGCHAINGO_POSTGRESQL")
	// if dsn == "" {
	// 	log.Fatal("Переменная окружения LANGCHAINGO_POSTGRESQL не установлена")
	// }

	// // Инициализация подключения к базе данных PostgreSQL
	// db, err := postgresql.NewPostgreSQL(dsn)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer db.Close()

	// // Создание SQL-цепочки

	// searchTool, err := duckduckgo.New(2, "")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// agentTools := []tools.Tool{
	// 	// &SQLDatabaseTool{
	// 	// 	// db: db.(*postgresql.PostgreSQL),

	// 	// },
	// 	searchTool,
	// }

	// mem := memory.NewConversationBuffer()
	// agent := agents.NewConversationalAgent(llm,
	// 	agentTools,
	// 	agents.WithMemory(mem), agents.WithMaxIterations(10))
	// executor := agents.NewExecutor(agent)
	// // Пример естественного языкового запроса
	// question := "Какая команда выиграла Blast Austin Major 2025? В качестве ответа дай просто название команды."

	// answer, err := chains.Run(context.Background(), executor, question)
	// fmt.Println(answer)
	// if err != nil {
	// 	log.Fatalf("Ошибка генерации ответа: %v", err)
	// }

}
