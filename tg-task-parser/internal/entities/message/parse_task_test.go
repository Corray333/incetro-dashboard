package message

import (
	"reflect"
	"testing"

	"github.com/corray333/tg-task-parser/internal/entities/task"
)

func TestParseTask_NoHashtag(t *testing.T) {
	tsk, err := ParseTask("Просто текст", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tsk != nil {
		t.Fatalf("expected nil task, got %#v", tsk)
	}
}

func TestParseTask_WithHashtag(t *testing.T) {
	tsk, err := ParseTask("сделать что-то #задача @user #golang", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tsk == nil {
		t.Fatalf("expected non-nil task")
	}
	// First rune should be uppercased according to ParseMessage
	if tsk.Text != "Сделать что-то" {
		t.Fatalf("unexpected text: %q", tsk.Text)
	}
	wantTags := []task.Tag{"задача", "golang"}
	if !reflect.DeepEqual(tsk.Hashtags, wantTags) {
		t.Fatalf("unexpected tags: got %#v want %#v", tsk.Hashtags, wantTags)
	}
	wantMentions := []task.Mention{"user"}
	if !reflect.DeepEqual(tsk.Mentions, wantMentions) {
		t.Fatalf("unexpected mentions: got %#v want %#v", tsk.Mentions, wantMentions)
	}
}