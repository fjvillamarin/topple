# PSX Language Grammar

PSX (Python Syntax eXtension) is a templating language that seamlessly blends Python's programming capabilities with HTML markup for creating dynamic web views. It allows developers to write UI components using familiar Python syntax directly within HTML structures, offering a development experience similar to JSX/TSX in the JavaScript ecosystem.

## Introduction

Key benefits of PSX:

- **Full Python Integration**: Use Python's complete feature set - control flow, data structures, functions, classes, and more - directly in templates
- **Component-Based Architecture**: Build modular, reusable UI components called "views"
- **Type Safety**: Full support for Python type annotations
- **Framework Integration**: Seamless integration with FastAPI, Flask, and other Python web frameworks
- **HTMX Support**: First-class support for building interactive UIs with minimal JavaScript

## Getting Started

### Installation

```bash
# Clone and build from source
git clone https://github.com/fjvillamarin/topple.git
cd topple
mise run build
# Copy binary to PATH or add bin/ to your PATH
cp bin/topple /usr/local/bin/
```

### Your First View

Create a file `hello.psx`:

```python
view HelloWorld():
    <div>
        <h1>Hello, World!</h1>
        <p>Welcome to Topple!</p>
    </div>
```

Compile it:

```bash
topple compile hello.psx
```

This generates `hello.py` which you can import and use in any Python application.

## View Definition

Views are the core building blocks in PSX:

```python
view ViewName(param1: Type1, param2: Type2 = default_value, *args, **kwargs):
    # View body with HTML and Python code
    <h1>Title</h1>
    
    if condition:
        <p>Conditional content</p>
    
    for item in items:
        <li>{item}</li>
```

- Views begin with the `view` keyword followed by a PascalCase name
- Support full Python parameter syntax including type annotations
- Can contain multiple root-level HTML elements
- Mix Python logic and HTML markup freely

## HTML Elements

### Element Types

1. **Block Elements** (Multi-line):
   ```python
   <div>
       <h1>Title</h1>
       <p>Content</p>
   </div>
   ```

2. **Inline Elements** (Single-line):
   ```python
   <span>Some text</span>
   ```

3. **Self-Closing Elements**:
   ```python
   <img src="image.jpg" />
   <input type="text" name="username" />
   ```

### Text Content

Text content in elements follows Python's expression rules:

1. **String Literals**: Direct Python strings
   ```python
   <div>
       "This is a plain string"
       'Single quotes work too'
   </div>
   ```

2. **F-Strings**: For dynamic content
   ```python
   view Greeting(name: str):
       <div>
           f"Hello, {name}!"
           f"You have {len(name)} letters in your name"
   </div>
   ```

3. **Expressions as Text**: Must be wrapped in f-strings
   ```python
   view ShowData(value: int, items: list):
       <div>
           f"{value * 2}"              # Expressions
           f"{len(items)} items"       # Function calls
           f"{items[0].upper()}"       # Method calls
       </div>
   ```

## Attributes

HTML attributes support three forms:

1. **Static Strings**:
   ```python
   <div class="container" id="main">
   ```

2. **Dynamic Expressions** (in curly braces):
   ```python
   <input type="checkbox" checked={is_active} />
   <div data-count={item_count}>
   ```

3. **F-String Attributes**:
   ```python
   <div class=f"item {'active' if selected else ''}" 
        id=f"item-{item_id}">
   ```

## Python Integration

### Control Flow

Use Python's control structures directly:

```python
view TodoList(todos: list[Todo]):
    <ul>
        for todo in todos:
            <li class={"done" if todo.completed else ""}>
                {todo.title}
                if todo.due_date:
                    <span class="due-date">
                        f"Due: {todo.due_date}"
                    </span>
            </li>
        else:
            <li>No todos yet!</li>
    </ul>
```

### Expressions

Any Python expression can be used:

```python
view DataDisplay(data: dict):
    total = sum(data.values())
    average = total / len(data) if data else 0
    
    <div class="stats">
        <h3>Statistics</h3>
        <p>f"Total: {total}"</p>
        <p>f"Average: {average:.2f}"</p>
        <ul>
            for key, value in sorted(data.items()):
                <li>f"{key}: {value}"</li>
        </ul>
    </div>
```

## View Composition

Views can contain other views:

```python
view Card(title: str):
    <div class="card">
        <h3>{title}</h3>
        <div class="content">
            <slot />
        </div>
    </div>

view UserProfile(user: User):
    <Card title={user.name}>
        <p>f"Email: {user.email}"</p>
        <p>f"Joined: {user.joined_date}"</p>
    </Card>
```

## Slots

> **Note**: Slots within HTML elements work as shown below. However, passing nested content to **view elements** (e.g., `<Card>...</Card>`) is not yet supported and will produce a compilation error.

Slots allow flexible content injection:

```python
view Layout():
    <html>
        <head>
            <slot name="head">
                <title>Default Title</title>
            </slot>
        </head>
        <body>
            <header>
                <slot name="header">
                    <h1>Default Header</h1>
                </slot>
            </header>
            <main>
                <slot />  # Default slot
            </main>
        </body>
    </html>

view HomePage():
    <Layout>
        <title slot="head">Home - My Site</title>
        <nav slot="header">
            <a href="/">Home</a>
            <a href="/about">About</a>
        </nav>
        <h1>Welcome!</h1>
        <p>This is the home page.</p>
    </Layout>
```

## HTMX Integration

PSX has first-class support for HTMX attributes:

```python
view SearchBox():
    <div>
        <input 
            type="search" 
            name="q"
            hx-get="/search"
            hx-trigger="keyup changed delay:500ms"
            hx-target="#results"
            placeholder="Search..."
        />
        <div id="results"></div>
    </div>

view TodoItem(todo: Todo):
    <li>
        <span>{todo.title}</span>
        <button 
            hx-delete=f"/todos/{todo.id}"
            hx-target="closest li"
            hx-swap="outerHTML"
        >
            Delete
        </button>
    </li>
```

## Advanced Features

### Match Statements

Python 3.10+ match statements are fully supported:

```python
view StatusDisplay(status: Status):
    <div class="status">
        match status:
            case Status.PENDING:
                <span class="pending">Pending...</span>
            case Status.ACTIVE:
                <span class="active">Active</span>
            case Status.COMPLETED:
                <span class="completed">Completed</span>
            case _:
                <span class="unknown">Unknown</span>
    </div>
```

### Error Handling

Use try/except blocks:

```python
view SafeDataView(data_id: str):
    <div>
        try:
            data = load_data(data_id)
            <div class="data">
                <h2>{data.title}</h2>
                <p>{data.description}</p>
            </div>
        except DataNotFoundError:
            <div class="error">
                <p>Data not found</p>
            </div>
        except Exception as e:
            <div class="error">
                <p>f"Error: {str(e)}"</p>
            </div>
    </div>
```

## Best Practices

1. **Use Type Hints**: Always annotate view parameters for better IDE support and documentation
2. **Keep Views Focused**: Each view should have a single responsibility
3. **Leverage Composition**: Build complex UIs from smaller, reusable views
4. **Handle Edge Cases**: Use Python's control flow to handle empty states and errors
5. **Escape User Content**: The runtime automatically escapes dynamic content for security

## Examples

### Complete Todo App

```python
from typing import List
from dataclasses import dataclass
from datetime import datetime

@dataclass
class Todo:
    id: int
    title: str
    completed: bool = False
    created_at: datetime = field(default_factory=datetime.now)

view TodoApp(todos: List[Todo]):
    <div class="todo-app">
        <h1>My Todos</h1>
        
        <form hx-post="/todos" hx-target="#todo-list" hx-swap="beforeend">
            <input type="text" name="title" placeholder="Add a todo..." required />
            <button type="submit">Add</button>
        </form>
        
        <ul id="todo-list">
            for todo in todos:
                <TodoItem todo={todo} />
        </ul>
        
        if not todos:
            <p class="empty">No todos yet. Add one above!</p>
    </div>

view TodoItem(todo: Todo):
    <li class={"completed" if todo.completed else ""}>
        <input 
            type="checkbox" 
            checked={todo.completed}
            hx-patch=f"/todos/{todo.id}/toggle"
            hx-target="closest li"
            hx-swap="outerHTML"
        />
        <span>{todo.title}</span>
        <button 
            class="delete"
            hx-delete=f"/todos/{todo.id}"
            hx-target="closest li"
            hx-swap="outerHTML swap:500ms"
        >
            ×
        </button>
    </li>
```

### FastAPI Integration

```python
from fastapi import FastAPI, Form
from fastapi.responses import HTMLResponse

app = FastAPI()

todos = []

@app.get("/", response_class=HTMLResponse)
async def index():
    return TodoApp(todos=todos).render()

@app.post("/todos", response_class=HTMLResponse)
async def add_todo(title: str = Form(...)):
    todo = Todo(id=len(todos), title=title)
    todos.append(todo)
    return TodoItem(todo=todo).render()

@app.patch("/todos/{todo_id}/toggle", response_class=HTMLResponse)
async def toggle_todo(todo_id: int):
    todo = todos[todo_id]
    todo.completed = not todo.completed
    return TodoItem(todo=todo).render()

@app.delete("/todos/{todo_id}")
async def delete_todo(todo_id: int):
    todos.pop(todo_id)
    return ""
```

## Advanced Examples

### Python Integration with Decorators

PSX fully supports Python's decorator syntax:

```python
from functools import wraps

def cache_result(func):
    cache = {}
    @wraps(func)
    def wrapper(*args, **kwargs):
        key = str(args) + str(kwargs)
        if key not in cache:
            cache[key] = func(*args, **kwargs)
        return cache[key]
    return wrapper

@cache_result
def expensive_operation(n: int) -> int:
    return n ** 2 + n * 10

view DecoratorDemo(values: list):
    results = [expensive_operation(x) for x in values if x > 0]
    
    <div>
        <h1>Processed {len(results)} values</h1>
        for i, result in enumerate(results):
            <div>Result {i}: {result}</div>
    </div>
```

### FastAPI Route Integration

Views can be directly used as FastAPI route handlers:

```python
from fastapi import FastAPI, Request
from fastapi.responses import HTMLResponse

app = FastAPI()

@app.get("/", response_class=HTMLResponse)
view HomePage(request: Request):
    <html>
        <head>
            <title>Topple App</title>
        </head>
        <body>
            <h1>Welcome to Topple!</h1>
            <p>Hello, visitor from {request.client.host}</p>
        </body>
    </html>

@app.get("/products/{product_id}")
view ProductDetail(product_id: int):
    <div class="product">
        <h1>Product #{product_id}</h1>
        <p>Product details for ID: {product_id}</p>
    </div>
```

### Complex Control Flow

```python
view DataDashboard(data: dict, user_role: str):
    <div class="dashboard">
        match user_role:
            case "admin":
                <div class="admin-panel">
                    <h2>Admin Dashboard</h2>
                    for key, value in data.items():
                        <div class="metric">
                            <strong>{key}:</strong> {value}
                        </div>
                </div>
            case "user":
                <div class="user-panel">
                    <h2>User Dashboard</h2>
                    <p>Summary: {len(data)} items</p>
                </div>
            case _:
                <div class="guest-panel">
                    <p>Please log in to view data</p>
                </div>
        
        try:
            stats = calculate_stats(data)
            <div class="stats">
                <h3>Statistics</h3>
                <pre>{stats}</pre>
            </div>
        except Exception as e:
            <div class="error">
                Error calculating stats: {str(e)}
            </div>
    </div>
```

### Dynamic Forms with HTMX

```python
view ContactForm(csrf_token: str):
    <form 
        hx-post="/api/contact"
        hx-target="#form-result"
        hx-swap="outerHTML"
        class="contact-form"
    >
        <input type="hidden" name="csrf_token" value={csrf_token} />
        
        <div class="form-group">
            <label for="name">Name:</label>
            <input 
                type="text" 
                id="name" 
                name="name" 
                required
                hx-post="/api/validate/name"
                hx-trigger="blur"
                hx-target="#name-error"
            />
            <span id="name-error"></span>
        </div>
        
        <div class="form-group">
            <label for="email">Email:</label>
            <input 
                type="email" 
                id="email" 
                name="email" 
                required
            />
        </div>
        
        <button type="submit">Send Message</button>
        <div id="form-result"></div>
    </form>
```

### Component Libraries

Create reusable UI component libraries:

```python
# ui_components.psx

view Button(text: str, variant: str = "primary", disabled: bool = False):
    <button 
        class=f"btn btn-{variant}"
        disabled={disabled}
    >
        {text}
    </button>

view Card(title: str, *, children=None, footer=None):
    <div class="card">
        <div class="card-header">
            <h3>{title}</h3>
        </div>
        <div class="card-body">
            {render_child(children) if children else "No content"}
        </div>
        if footer:
            <div class="card-footer">
                {render_child(footer)}
            </div>
    </div>

view Alert(message: str, type: str = "info", dismissible: bool = True):
    <div class=f"alert alert-{type}" role="alert">
        {message}
        if dismissible:
            <button 
                type="button" 
                class="close" 
                aria-label="Close"
                onclick="this.parentElement.remove()"
            >
                <span aria-hidden="true">&times;</span>
            </button>
    </div>
```

### Real-time Updates with WebSockets

```python
view ChatRoom(room_id: str, ws_url: str):
    <div class="chat-room" data-room-id={room_id}>
        <div 
            id="messages"
            hx-ws=f"connect:{ws_url}/ws/{room_id}"
            class="messages-container"
        >
            <div class="message system">
                Connected to room {room_id}
            </div>
        </div>
        
        <form 
            hx-ws=f"send:{ws_url}/ws/{room_id}"
            class="message-form"
        >
            <input 
                type="text" 
                name="message" 
                placeholder="Type a message..."
                required
            />
            <button type="submit">Send</button>
        </form>
    </div>
```

## See Also

- [Architecture Guide](architecture.md) - Deep dive into compiler internals
- [CLI Reference](cli.md) - Command-line tool documentation
- [Runtime System](runtime.md) - Runtime library documentation
- [AST Reference](development/ast_reference.md) - Complete AST node reference
- [Examples](../examples/) - More example applications