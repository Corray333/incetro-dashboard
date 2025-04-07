package task

import (
	"reflect"
	"sort"
	"testing"
)

func sortTaskFields(task *Task) {
	sort.Slice(task.Hashtags, func(i, j int) bool {
		return task.Hashtags[i] < task.Hashtags[j]
	})
	sort.Slice(task.Mentions, func(i, j int) bool {
		return task.Mentions[i] < task.Mentions[j]
	})
}

func TestTaskFromMessage(t *testing.T) {
	tests := []struct {
		name      string
		mainText  string
		replyText string
		wantTask  *Task
	}{
		{
			name:      "basic task with hashtag and mention from main text",
			mainText:  "Нужно исправить баг #golang @user1",
			replyText: "",
			wantTask: &Task{
				Text:     "Нужно исправить баг",
				Hashtags: []Hashtag{"golang"},
				Mentions: []Mention{"user1"},
			},
		},
		{
			name:      "task from reply text",
			mainText:  "комментарий @user2",
			replyText: "Нужно исправить баг #golang #задача @user1",
			wantTask: &Task{
				Text:     "Нужно исправить баг",
				Hashtags: []Hashtag{"golang", "задача"},
				Mentions: []Mention{"user1", "user2"},
			},
		},
		{
			name:      "no hashtags or mentions",
			mainText:  "Просто задача без тегов",
			replyText: "",
			wantTask: &Task{
				Text:     "Просто задача без тегов",
				Hashtags: nil,
				Mentions: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTask, err := TaskFromMessage(tt.mainText, tt.replyText)
			if err != nil {
				t.Fatalf("TaskFromMessage() returned unexpected error: %v", err)
			}

			sortTaskFields(gotTask)
			sortTaskFields(tt.wantTask)

			if !reflect.DeepEqual(gotTask, tt.wantTask) {
				t.Errorf("TaskFromMessage() = %+v,\nwant %+v", gotTask, tt.wantTask)
			}
		})
	}
}
