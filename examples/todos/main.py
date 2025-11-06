from fastapi import FastAPI, Form, Header
from fastapi.responses import HTMLResponse
from typing import Optional

# Import compiled PSX views
from examples.todos.views import TodoPage, TodoList

app = FastAPI()

# In-memory storage (use a database in production!)
todos_storage = []
next_id = 1


@app.get("/", response_class=HTMLResponse)
async def home():
    """Render the full todo page"""
    page = TodoPage(todos=todos_storage)
    return page.render()


@app.post("/todos", response_class=HTMLResponse)
async def add_todo(
    todo: str = Form(...),
    hx_request: Optional[str] = Header(None)
):
    """Add a new todo and return the updated list"""
    global next_id

    # Add the new todo
    new_todo = {
        "id": next_id,
        "text": todo
    }
    todos_storage.append(new_todo)
    next_id += 1

    # Return just the todo list fragment (HTMX will swap it)
    todo_list = TodoList(todos=todos_storage)
    return todo_list.render()


@app.delete("/todos/{todo_id}", response_class=HTMLResponse)
async def delete_todo(
    todo_id: int,
    hx_request: Optional[str] = Header(None)
):
    """Delete a todo and return the updated list"""
    global todos_storage

    # Filter out the todo with matching id
    todos_storage = [t for t in todos_storage if t.get("id") != todo_id]

    # Return the updated todo list fragment
    todo_list = TodoList(todos=todos_storage)
    return todo_list.render()


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="127.0.0.1", port=8000)
