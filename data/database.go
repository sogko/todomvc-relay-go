package data

import "fmt"

// Mock authenticated ID
const ViewerId = "me"

// Model structs
type Todo struct {
	ID       string `json:"id"`
	Text     string `json:"text"`
	Complete bool   `json:"complete"`
}

type User struct {
	ID string `json:"id"`
}

// Mock data
var viewer = &User{ViewerId}
var usersById = map[string]*User{
	ViewerId: viewer,
}
var todosById = map[string]*Todo{}
var todoIdsByUser = map[string][]string{
	ViewerId: []string{},
}
var nextTodoId = 0

// Data methods

func AddTodo(text string, complete bool) string {
	todo := &Todo{
		ID:       fmt.Sprintf("%v", nextTodoId),
		Text:     text,
		Complete: complete,
	}
	nextTodoId = nextTodoId + 1

	todosById[todo.ID] = todo
	todoIdsByUser[ViewerId] = append(todoIdsByUser[ViewerId], todo.ID)

	return todo.ID
}

func GetTodo(id string) *Todo {
	if todo, ok := todosById[id]; ok {
		return todo
	}
	return nil
}

func GetTodos(status string) []*Todo {
	todos := []*Todo{}
	for _, todoId := range todoIdsByUser[ViewerId] {
		if todo := GetTodo(todoId); todo != nil {

			switch status {
			case "completed":
				if todo.Complete {
					todos = append(todos, todo)
				}
			case "incomplete":
				if !todo.Complete {
					todos = append(todos, todo)
				}
			case "any":
				fallthrough
			default:
				todos = append(todos, todo)
			}
		}
	}
	return todos
}

func GetUser(id string) *User {
	if user, ok := usersById[id]; ok {
		return user
	}
	return nil
}

func GetViewer() *User {
	return GetUser(ViewerId)
}

func ChangeTodoStatus(id string, complete bool) {
	todo := GetTodo(id)
	if todo == nil {
		return
	}
	todo.Complete = complete
}

func MarkAllTodos(complete bool) []string {
	changedTodoIds := []string{}
	for _, todo := range GetTodos("any") {
		if todo.Complete != complete {
			todo.Complete = complete
			changedTodoIds = append(changedTodoIds, todo.ID)
		}
	}
	return changedTodoIds
}

func RemoveTodo(id string) {

	updatedTodoIdsForUser := []string{}
	for _, todoId := range todoIdsByUser[ViewerId] {
		if todoId != id {
			updatedTodoIdsForUser = append(updatedTodoIdsForUser, todoId)
		}
	}
	todoIdsByUser[ViewerId] = updatedTodoIdsForUser
	delete(todosById, id)

}

func RemoveCompletedTodos() []string {
	todosIdRemoved := []string{}
	for _, completedTodo := range GetTodos("completed") {
		RemoveTodo(completedTodo.ID)
		todosIdRemoved = append(todosIdRemoved, completedTodo.ID)
	}
	return todosIdRemoved
}

func RenameTodo(id string, text string) {
	todo := GetTodo(id)
	if todo != nil {
		todo.Text = text
	}
}

func TodosToSliceInterface(todos []*Todo) []interface{} {
	todosIFace := []interface{}{}
	for _, todo := range todos {
		todosIFace = append(todosIFace, todo)
	}
	return todosIFace
}
