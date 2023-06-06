package main

import (
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/tekqer/hr"
)

var (
	store  = sync.Map{}
	nextid atomic.Int32
)

type Todo struct {
	Id    int     `json:"id"`
	Title string  `json:"title" form:"title"`
	Time  hr.Time `json:"time"`
}

func main() {
	r := hr.Default()

	r.GET("/todo/:id", func(c *hr.Ctx) error {
		id, err := strconv.Atoi(c.Vars().Get("id"))
		if err != nil {
			return hr.BadRequest("bad id")
		}
		val, ok := store.Load(id)
		if !ok {
			return hr.NotFound("no record with id %d", id)
		}
		return c.JSON(val)
	})

	r.DELETE("/todo/:id", func(c *hr.Ctx) error {
		id, err := strconv.Atoi(c.Vars().Get("id"))
		if err != nil {
			return hr.BadRequest("bad id")
		}
		store.Delete(id)
		return nil
	})

	r.POST("/todo", func(c *hr.Ctx) error {
		var todo Todo
		if err := c.Bind(&todo); err != nil {
			return err
		}
		todo.Id = int(nextid.Add(1))
		todo.Time = hr.Time(time.Now())
		store.Store(todo.Id, &todo)
		return nil
	})

	r.GET("/todo", func(c *hr.Ctx) error {
		todos := []*Todo{}
		store.Range(func(key, value any) bool {
			todos = append(todos, value.(*Todo))
			return true
		})
		return c.JSON(todos)
	})

	http.ListenAndServe(":8080", r)
}
