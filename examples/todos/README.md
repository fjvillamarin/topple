# PSX + HTMX Todo List Example

A practical demonstration of server-side rendering with interactivity using PSX (Python Syntax eXtension) and HTMX.

## What This Demonstrates

- **PSX Syntax**: HTML-like syntax in Python with loops and conditionals
- **Server-Side Rendering**: All rendering happens on the server
- **HTMX Interactivity**: Dynamic updates without writing JavaScript
- **Multiple Endpoints**: POST for adding, DELETE for removing
- **Partial Updates**: Only the todo list updates, not the entire page
- **Zero Client-Side JavaScript**: All logic runs server-side

## Features

- Add new todos via form submission
- Delete todos with one click
- Empty state when no todos exist
- Beautiful gradient UI with Tailwind CSS
- Fast, responsive interactions via HTMX

## How It Works

### PSX Views (`views.psx`)

Two views are defined:

1. **`TodoPage(todos: list)`** - Full HTML page with:
   - HTMX script loaded from CDN
   - Tailwind CSS for styling
   - Form with `hx-post` to add todos
   - Container for todo list

2. **`TodoList(todos: list)`** - Just the list fragment:
   - Loops over todos to render each item
   - Delete button with `hx-delete`
   - Empty state when list is empty

### FastAPI Server (`main.py`)

Three endpoints:

- **`GET /`** - Returns full `TodoPage` with current todos
- **`POST /todos`** - Adds todo, returns `TodoList` fragment
- **`DELETE /todos/{id}`** - Removes todo, returns `TodoList` fragment

HTMX automatically swaps the returned HTML into the `#todo-list` container.

## Running the Example

### 1. Compile PSX to Python (if not already done)

```bash
./bin/topple compile examples/todos/views.psx
```

This generates `examples/todos/views.py` from the PSX source.

### 2. Start the Server

Using Poetry (recommended):
```bash
poetry run uvicorn examples.todos.main:app --reload
```

Or using Python directly:
```bash
python -m uvicorn examples.todos.main:app --reload
```

### 3. Open in Browser

Navigate to: http://127.0.0.1:8000

## HTMX Patterns Used

### Form Submission
```html
<form hx-post="/todos" hx-target="#todo-list" hx-swap="innerHTML">
```
- `hx-post`: POST form data to `/todos` endpoint
- `hx-target`: Update the `#todo-list` element
- `hx-swap="innerHTML"`: Replace the inner HTML

### Delete Button
```html
<button hx-delete="/todos/{id}" hx-target="#todo-list">
```
- `hx-delete`: Send DELETE request
- Dynamic URL with todo ID
- Same target for consistent updates

## Code Structure

```
examples/todos/
├── __init__.py      # Python package marker
├── main.py          # FastAPI server with endpoints
├── views.psx        # PSX source with view definitions
├── views.py         # Generated Python (don't edit manually)
└── README.md        # This file
```

## PSX Features Demonstrated

### Conditionals in HTML
```python
if len(todos) == 0:
    <div>No todos yet!</div>
else:
    <ul>...</ul>
```

### Loops in HTML
```python
for todo in todos:
    <li>{todo.get("text", "")}</li>
```

### Expression Interpolation
```python
<button hx-delete="/todos/{todo.get('id', '')}">Delete</button>
```

### Safe Dictionary Access
```python
{todo.get("text", "")}    # Won't raise KeyError
{todo.get('id', '')}       # Returns empty string if missing
```

## Architecture Notes

### In-Memory Storage
This example uses a simple list for storage:
```python
todos_storage = []
```

**In production, you should:**
- Use a database (PostgreSQL, SQLite, etc.)
- Add user authentication
- Implement proper sessions
- Add validation and sanitization

### HTMX Request Detection
The `hx_request` header can differentiate HTMX from normal requests:
```python
async def add_todo(hx_request: Optional[str] = Header(None)):
    if hx_request:
        # Return just the fragment
    else:
        # Return full page
```

Currently, we always return fragments since HTMX makes all requests after initial load.

## Extending This Example

Ideas for enhancements:

1. **Mark as Complete**: Add a checkbox to toggle todo completion
2. **Edit Todos**: Click to edit text inline
3. **Persistence**: Save to SQLite database
4. **Filtering**: Show all/active/completed todos
5. **Search**: Filter todos by text
6. **Priorities**: Add color-coded priority levels

## Why PSX + HTMX?

This combination offers several advantages:

- **Simpler Stack**: No Node.js, no build step, no JavaScript framework
- **Better Performance**: Server renders HTML faster than client-side hydration
- **Easier Debugging**: Python stack traces vs JavaScript console errors
- **Type Safety**: Full Python type hints and IDE support
- **Progressive Enhancement**: Works without JavaScript (gracefully degrades)

## Learn More

- **Topple/PSX**: https://github.com/topple-lang/topple
- **HTMX**: https://htmx.org
- **FastAPI**: https://fastapi.tiangolo.com

## License

Same as the Topple project.
