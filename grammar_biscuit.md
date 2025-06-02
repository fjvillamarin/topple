# Introduction

Biscuit is a templating language designed to seamlessly blend Python's programming capabilities with HTML for creating dynamic web views. It allows developers to write UI components using familiar Python syntax directly within HTML-like structures, offering a development experience similar to JSX/TSX in the JavaScript ecosystem.

Key benefits of using Biscuit include:

-   **Leveraging Python's full power within templates**: Utilize Python's control flow (if/else, loops), data structures (lists, dictionaries), functions, and even classes directly in your views.
-   **Creating reusable UI components (views)**: Build modular and maintainable user interfaces by encapsulating markup and logic into reusable views.
-   **Direct integration with FastAPI**: Biscuit views can be used directly as FastAPI route handlers, simplifying the development of web applications.
-   **Support for HTMX**: Enhance interactivity and create modern user experiences with minimal JavaScript, thanks to built-in support for HTMX attributes.

# Getting Started: A Simple Example

This tutorial will guide you through creating a basic "Hello, World!" application using Biscuit with FastAPI. We'll define a simple view, set up a FastAPI server, and render our view.

The complete code for this example can be found in `examples/biscuit/01_hello_world/01_hello_world.bsct`.

Let's break down the code step-by-step:

**1. Define a Biscuit View**

A Biscuit view is defined using the `view` keyword, followed by a name (typically PascalCase) and parentheses for parameters (if any). Inside the view, you write HTML markup mixed with Python expressions.

```python
view HelloWorld():
    <div>
        <h1>Hello, World!</h1>
        <p>This is my first Biscuit view.</p>
    </div>
```
In this example, `HelloWorld` is a simple view that renders a heading and a paragraph.

**2. Import Necessary Libraries**

To run a Biscuit application with FastAPI, you need to import a few things:
- `FastAPI` from `fastapi` for creating the web application.
- `HTMLResponse` from `fastapi.responses` to indicate that your endpoint will return HTML.
- `uvicorn` to serve your FastAPI application.

```python
from fastapi import FastAPI
from fastapi.responses import HTMLResponse
import uvicorn
```

**3. Create a FastAPI Application Instance**

Instantiate `FastAPI`. You can optionally give it a title.

```python
app = FastAPI(title="Biscuit Example: Hello World")
```

**4. Create a Route and Render the View**

Define a route using FastAPI's decorators (e.g., `@app.get("/")`). In the route function, you render your Biscuit view by calling its `.render()` method. The result of `.render()` is an HTML string, which should be returned wrapped in an `HTMLResponse`.

```python
@app.get("/", response_class=HTMLResponse)
async def index():
    return HelloWorld().render()
```
Here, when a user visits the root path (`/`), the `index` function is called, which renders the `HelloWorld` view and returns it as HTML.

**5. Run the Application**

The standard Python `if __name__ == "__main__":` block is used to run the application using `uvicorn`.

```python
if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8000)
```
This makes your application accessible at `http://0.0.0.0:8000` (or `http://localhost:8000`).

**Complete Example (`01_hello_world.bsct`)**

Here's the full code putting all the pieces together:

```python
from fastapi import FastAPI
from fastapi.responses import HTMLResponse
import uvicorn

# 1. Define a Biscuit View
view HelloWorld():
    <div>
        <h1>Hello, World!</h1>
        <p>This is my first Biscuit view.</p>
    </div>

# 3. Create a FastAPI Application Instance
app = FastAPI(title="Biscuit Example: Hello World")

# 4. Create a Route and Render the View
@app.get("/", response_class=HTMLResponse)
async def index():
    return HelloWorld().render()

# 5. Run the Application
if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8000)
```

This simple example demonstrates the core workflow of a Biscuit application. To explore more advanced features and complex use cases, please check out the other examples in the `examples/` directory.

# Biscuit Language Grammar

Biscuit is a templating language that combines Python syntax with HTML-like markup to create reusable UI views. This document explains the core grammar and syntax of the Biscuit language.

## Core Elements

### View Definition

```python
view ViewName(param1: Type1, param2: Type2 = default_value, *args, **kwargs):
    # View body with multiple HTML statements and Python code
    <h1>Title</h1>
    <p>First paragraph</p>
    
    if condition:
        <p>Conditional content</p>
    
    <div>
        <p>More content</p>
    </div>
```

- Views begin with the `view` keyword followed by a PascalCase name
- Support Python type annotations for parameters
- Allow default values for parameters
- Support variable arguments (*args) and keyword arguments (**kwargs)
- Body contains a mix of Python logic and multiple HTML statements
- Unlike components in other frameworks, a view can contain multiple root-level HTML elements

### HTML-like Tags

Three types of HTML tag structures are supported:

1. **Multiline HTML tags**:
   ```html
   <div>
       <h1>Title</h1>
       <p>Content</p>
   </div>
   ```

2. **Single-line HTML tags**:
   ```html
   <span>Some text</span>
   ```

3. **Self-closing tags**:
   ```html
   <img src="image.jpg" />
   ```

### Attributes

HTML tags can have two types of attributes:

1. **Value attributes**: `<div class="container" id="main">`
2. **Boolean attributes**: `<button disabled>`

### String Interpolation

Python expressions can be embedded in HTML content using curly braces:

```html
<h1>Hello, {user.name}!</h1>
<p>Your account balance is {format_currency(account.balance)}</p>
```

### Control Flow

Biscuit supports Python-style control structures:

#### Conditional Rendering
```python
if condition:
    <div>Content shown if condition is true</div>
elif other_condition:
    <div>Alternative content</div>
else:
    <div>Fallback content</div>
```

#### Loops
```python
for item in items:
    <div>
        <h1>{item.title}</h1>
        <p>{item.description}</p>
    </div>
```

```python
while condition:
    <div>Repeated content</div>
```

#### Early Returns
```python
if not items:
    <div>No items found</div>
    return
```

## View Composition

Biscuit uses a JSX-like syntax for view composition with Vue-inspired slots:

```python
view UserAvatar(user, size="medium"):
    <div class="avatar {size}">
        if user.profile_image:
            <img src="{user.profile_image}" alt="{user.name}" />
        else:
            <div class="avatar-placeholder">
                <span>{user.name[0].upper()}</span>
            </div>
    </div>

view UserDetails(user):
    <div class="user-details">
        <h2>{user.name}</h2>
        <p class="user-email">{user.email}</p>
        <slot>
            <p>No additional details provided</p>
        </slot>
    </div>

view UserCard(user):
    <div class="user-card">
        <div class="card-header">
            <slot name="header">
                <h3>{user.name}</h3>
            </slot>
        </div>
        <div class="card-body">
            <slot>
                <p>No content provided</p>
            </slot>
        </div>
        <div class="card-footer">
            <slot name="footer" />
        </div>
    </div>

view UserProfile(user):
    <h1>Profile for {user.name}</h1>
    <UserAvatar user={user} size="large" />
    <UserDetails user={user}>
        <p>Member since {user.join_date}</p>
    </UserDetails>
    
    <UserCard user={user}>
        <div slot="header">
            <h2>{user.name} - {user.role}</h2>
        </div>
        <p>User bio and details go here</p>
        <div slot="footer">
            <button>Edit Profile</button>
        </div>
    </UserCard>
```

Key features of view composition:

- Views are used as HTML-like tags with PascalCase names
- Parameters are passed as attributes using curly brace syntax for dynamic values
- Static string values can be passed without curly braces: `title="My Title"`
- Boolean parameters can be passed using flag syntax: `disabled`
- Views can have children, which will be available to the receiving view
- Views can be nested arbitrarily deep
- Biscuit intelligently differentiates between HTML tags and views:
  - PascalCase tags (`<UserAvatar>`) are treated as views if in scope
  - If a PascalCase tag is not in scope, it's treated as a regular HTML element
  - lowercase tags (`<div>`) are always treated as HTML elements

### Multiple Root Elements

Unlike components in most UI frameworks, Biscuit views can have multiple root-level elements:

```python
view UserGreeting(user):
    <h1>Hello, {user.name}!</h1>
    <p>Welcome back to our platform.</p>
    
    if user.is_new:
        <div class="tutorial">
            <p>Here's how to get started...</p>
        </div>
```

### Slots System

Biscuit provides a Vue.js-inspired slot system for flexible content distribution. Slots allow you to pass markup and other views into a component from its parent, enabling greater component reusability and composition.

Within the component defining the slots (the child), you use the `<slot>` element. To pass content into these slots from the parent component, you can use direct children (for the default slot) or the `<template>` tag with a `slot` attribute for named slots. Alternatively, any HTML element can have a `slot` attribute to direct its content to a specific named slot.

##### Basic Default Slot

Content passed directly as children to a component, without a `slot` attribute, populates the default (unnamed) slot.

```python
# Defining a view with a default slot
view Card():
    <div class="card">
        <div class="card-header">
            <slot name="header"> <!-- Named slot for header -->
                <h4>Default Header</h4>
            </slot>
        </div>
        <div class="card-body">
            <slot /> <!-- Default slot for main content -->
        </div>
    </div>

# Using the view with default slot content
<Card>
    <template slot="header">
        <h1>My Custom Card Header</h1>
    </template>

    <!-- This content goes into the default slot -->
    <p>This is the main content of the card.</p>
    <p>It can contain multiple elements.</p>
</Card>

# Alternative usage for default slot (implicit)
<Card>
    <!-- This content also goes into the default slot -->
    <p>This is the main content of the card.</p>
    <p>It can contain multiple elements.</p>
    <!-- To also use the named "header" slot here, you'd add a <template slot="header"> -->
</Card>
```

##### Named Slots

For more complex layouts, named slots allow multiple content areas. Use the `name` attribute on the `<slot>` element to define a named slot. To provide content, use the `<template>` tag with a `slot` attribute matching the slot's name, or add a `slot` attribute to any HTML element.

```python
# Defining a view with named slots
view PageLayout():
    <div class="layout">
        <header>
            <slot name="header" />
        </header>
        <main>
            <slot /> <!-- Default slot -->
        </main>
        <footer>
            <slot name="footer" />
        </footer>
    </div>

# Using named slots with <template>
<PageLayout>
    <template slot="header">
        <h1>Main Page Title</h1>
    </template>
    
    <!-- Content for the default slot -->
    <p>This is the primary content of the page.</p>
    
    <template slot="footer">
        <p>&copy; 2024 My Application</p>
    </template>
</PageLayout>

# Using named slots with direct HTML elements having a 'slot' attribute
<PageLayout>
    <h1 slot="header">Alternative Page Title</h1>

    <p>Main content for the default slot.</p>

    <div slot="footer" class="footer-content">
        <p>&copy; 2024 My Application - Alternative Footer</p>
    </div>
</PageLayout>
```
Using `<template slot="name">` is generally preferred for clarity, especially for complex slot content, as it doesn't introduce an extra HTML element into the slot's structure unless that element is part of the content itself. Using `<div slot="name">` (or any other tag) is a more direct way if that element is the intended wrapper.

##### Fallback Content

Slots can provide fallback (default) content that is rendered if no content is provided for that slot by the parent.

```python
# Defining fallback content in slots
view Alert(type: str = "info"):
    <div class="alert alert-{type}">
        <div class="alert-icon">
            <slot name="icon">
                <!-- Fallback icon content -->
                <img src="/icons/default-{type}-icon.png" alt="Icon" />
            </slot>
        </div>
        <div class="alert-content">
            <slot>
                <!-- Fallback default slot content -->
                <p>This is a default alert message.</p>
            </slot>
        </div>
    </div>

# Usage:
# 1. Providing custom content for the default slot, icon slot uses fallback
<Alert type="warning">
    <p>This is a custom warning message.</p>
</Alert>

# 2. Providing custom content for both icon and default slots
<Alert type="error">
    <template slot="icon">
        <img src="/icons/custom-error-icon.png" alt="Error Icon" />
    </template>
    <p>A critical error has occurred!</p>
</Alert>

# 3. Using fallback content for all slots
<Alert type="success" />
```

##### Conditional Slot Rendering

You can conditionally render elements based on whether a slot has been provided with content from the parent, using the `has_slot("slot_name")` function. For the default slot, use `has_slot("default")` or simply `has_slot()`.

```python
view Panel(title: str):
    <div class="panel">
        <div class="panel-header">
            <h3>{title}</h3>
            if has_slot("actions"): # Check for a named slot "actions"
                <div class="panel-actions">
                    <slot name="actions" />
                </div>
        </div>
        <div class="panel-body">
            <slot /> <!-- Default slot -->
        </div>
        
        if has_slot("footer"): # Check for a named slot "footer"
            <div class="panel-footer">
                <slot name="footer" />
            </div>
        else:
            <div class="panel-footer text-muted">
                <p>No footer content provided.</p>
            </div>
    </div>

# Usage:
# 1. Panel with default content and footer
<Panel title="My Panel">
    <p>This is the main content of the panel.</p>
    <template slot="footer">
        <button>Save</button>
    </template>
</Panel>

# 2. Panel with actions, default content, but no footer (will show "No footer content provided")
<Panel title="Another Panel">
    <template slot="actions">
        <button>Edit</button>
        <button>Delete</button>
    </template>
    <p>Some information here.</p>
</Panel>
```
The `has_slot()` function checks if any content (even empty content like `<template slot="footer"></template>`) has been passed to the slot.

## Key Syntax Elements

1. **View definition**: `view Name(params...):`
2. **Tag opening**: `<tagname attributes...>`
3. **Tag closing**: `</tagname>`
4. **Self-closing tag**: `<tagname attributes.../>`
5. **Expression interpolation**: `{expression}`
6. **View usage**: `<ViewName param1={value1}>...</ViewName>`
7. **Self-closing view**: `<ViewName param1={value1} />`

## Python Integration

Biscuit provides full access to Python features including:
- Variable assignments
- Function calls
- Type annotations
- List/dictionary comprehensions
- Lambda expressions
- Python operators and expressions

## Compilation Process

Biscuit files (`.bsct`) are compiled to Python code through:
1. Parsing the Biscuit syntax tree using tree-sitter
2. Transforming views into Python classes
3. Creating render methods that output HTML strings
4. Handling interpolation via Python f-strings
5. Resolving view references based on imports and scope

# HTMX Integration

Biscuit provides first-class support for HTMX, enabling rich, dynamic interactions with minimal JavaScript:

```python
view TodoList(todos, user_id):
    <div>
        <h1>Todo List</h1>
        <ul id="todo-list">
            for todo in todos:
                <li class={"completed" if todo.completed else ""}>
                    <span>{todo.title}</span>
                    <button 
                        hx-delete="/todos/{todo.id}"
                        hx-target="closest li"
                        hx-swap="outerHTML"
                        hx-confirm="Are you sure you want to delete this task?">
                        Delete
                    </button>
                </li>
        </ul>
        
        <div class="add-todo">
            <form hx-post="/todos" hx-target="#todo-list" hx-swap="beforeend">
                <input type="hidden" name="user_id" value="{user_id}" />
                <input type="text" name="title" placeholder="New task..." />
                <button type="submit">Add</button>
            </form>
        </div>
    </div>
```

### Creating HTMX Partial Views

You can create views that serve as HTMX endpoints, returning only the HTML fragments needed for partial page updates:

```python
view TodoItem(todo):
    <li class={"completed" if todo.completed else ""}>
        <span>{todo.title}</span>
        <input 
            type="checkbox" 
            checked={todo.completed}
            hx-post="/todos/{todo.id}/toggle"
            hx-target="closest li"
            hx-swap="outerHTML"
        />
        <button 
            hx-delete="/todos/{todo.id}"
            hx-target="closest li"
            hx-swap="outerHTML">
            Delete
        </button>
    </li>

# This view would be called from a server endpoint
view NewTodoResponse(todo):
    <TodoItem todo={todo} />
```

### Dynamic Loading with HTMX

```python
view LazyLoadingSection(url, trigger="revealed"):
    <div 
        hx-get="{url}" 
        hx-trigger="{trigger}"
        hx-swap="innerHTML"
        class="loading-container">
        <p>Loading...</p>
    </div>
```

### Form Processing with HTMX

HTMX works great for forms with validation and dynamic responses:

```python
view ContactForm():
    <div>
        <form hx-post="/contact" hx-swap="outerHTML">
            <div class="form-group">
                <label for="name">Name</label>
                <input type="text" id="name" name="name" required />
            </div>
            
            <div class="form-group">
                <label for="email">Email</label>
                <input type="email" id="email" name="email" required />
            </div>
            
            <div class="form-group">
                <label for="message">Message</label>
                <textarea id="message" name="message" required></textarea>
            </div>
            
            <button type="submit">Send Message</button>
        </form>
    </div>
```

### Form Validation Response

```python
view FormValidationError(errors):
    <div class="error-banner">
        <h3>Please correct the following errors:</h3>
        <ul>
            for field, error in errors.items():
                <li>{field}: {error}</li>
        </ul>
    </div>
```

### Infinite Scroll with HTMX

```python
view ArticleList(articles, page=1):
    for article in articles:
        <article class="article-card">
            <h2>{article.title}</h2>
            <p>{article.summary}</p>
            <a href="/articles/{article.id}">Read more</a>
        </article>
    
    if articles:
        <div 
            hx-get="/articles?page={page + 1}" 
            hx-trigger="revealed"
            hx-swap="afterend"
            class="loading-indicator">
            <span>Loading more articles...</span>
        </div>
```

### Active Search with HTMX

```python
view SearchInterface():
    <div class="search-container">
        <input 
            type="search" 
            name="q"
            placeholder="Search products..."
            hx-get="/search"
            hx-trigger="keyup changed delay:500ms, search"
            hx-target="#search-results"
            hx-indicator=".spinner"
        />
        
        <div class="spinner" style="display:none;">Searching...</div>
        <div id="search-results"></div>
    </div>
```

### Tabs with HTMX

```python
view TabInterface(tabs):
    <div class="tabs">
        <div class="tab-headers">
            for i, tab in enumerate(tabs):
                <button 
                    hx-get="/tabs/{tab.id}"
                    hx-target="#tab-content"
                    class={"active" if i == 0 else ""}>
                    {tab.title}
                </button>
        </div>
        
        <div id="tab-content">
            # Initial tab content will load here
            if tabs:
                <div>{tabs[0].content}</div>
        </div>
    </div>
```

# Error Management

Biscuit supports standard Python exception handling within views:

```python
view UserProfile(user_id):
    try:
        user = get_user_by_id(user_id)
        <div class="profile">
            <h1>{user.name}</h1>
            <p>{user.email}</p>
        </div>
    except UserNotFoundError:
        <div class="error-message">
            <h2>User not found</h2>
            <p>The requested user profile could not be found.</p>
        </div>
    except DatabaseError as e:
        <div class="error-message">
            <h2>Database error</h2>
            <p>There was a problem fetching the user profile: {str(e)}</p>
            <a href="/">Return to home page</a>
        </div>
    finally:
        log_profile_access_attempt(user_id)
```

Error handling best practices in views:
- Use try/except blocks for operations that might fail
- Provide appropriate error messages and fallback UI
- Handle specific exceptions with targeted error views
- Use the finally block for cleanup operations
- Avoid swallowing exceptions without proper handling

# FastAPI Integration

Biscuit seamlessly integrates with FastAPI to create HTML responses:

## Basic Route Handlers

Views can be directly used as FastAPI route handlers and automatically return HTMLResponse:

```python
from fastapi import FastAPI, Request

app = FastAPI()

@app.get("/")
view HomePage(request: Request):
    <html>
        <head>
            <title>Biscuit App</title>
        </head>
        <body>
            <h1>Welcome to Biscuit!</h1>
            <p>Hello, visitor from {request.client.host}</p>
        </body>
    </html>
```

## Path Parameters

FastAPI path parameters work directly with Biscuit views:

```python
@app.get("/products/{product_id}")
view ProductDetail(product_id: int, db: Session = Depends(get_db)):
    product = db.query(Product).get(product_id)
    
    <div class="product">
        <h1>{product.name}</h1>
        <p class="price">${product.price}</p>
        <div class="description">{product.description}</div>
    </div>
```

## Query Parameters

Query parameters are received as function arguments:

```python
@app.get("/search")
view SearchResults(
    q: str,
    category: Optional[str] = None,
    sort: str = "relevance",
    page: int = 1,
    db: Session = Depends(get_db)
):
    results = search_products(db, q, category, sort, page)
    
    <div class="search-results">
        <h1>Search Results for "{q}"</h1>
        
        if category:
            <p>In category: {category}</p>
            
        <p>Sorted by: {sort}</p>
        <p>Page: {page}</p>
        
        if not results:
            <p>No results found</p>
        else:
            <div class="results">
                for product in results:
                    <ProductCard product={product} />
            </div>
    </div>
```

## Form Data

Handling POST requests with form data:

```python
from fastapi import Form, Depends

@app.post("/contact")
view ContactForm(
    request: Request,
    name: str = Form(...),
    email: str = Form(...),
    message: str = Form(...),
    db: Session = Depends(get_db)
):
    # Process the form data
    contact_entry = ContactEntry(name=name, email=email, message=message)
    db.add(contact_entry)
    db.commit()
    
    # Show success message
    <div class="contact-success">
        <h1>Thank you, {name}!</h1>
        <p>Your message has been received. We'll respond to {email} shortly.</p>
        <a href="/">Return to homepage</a>
    </div>
```

## Request Body

Working with JSON request bodies:

```python
from pydantic import BaseModel

class UserCreate(BaseModel):
    username: str
    email: str
    password: str

@app.post("/users/new")
view CreateUser(
    user: UserCreate,
    db: Session = Depends(get_db)
):
    # Create user in database
    db_user = User(**user.dict(exclude={"password"}))
    db_user.hashed_password = hash_password(user.password)
    db.add(db_user)
    db.commit()
    
    <div class="user-created">
        <h1>User Created</h1>
        <p>Username: {db_user.username}</p>
        <p>Email: {db_user.email}</p>
    </div>
```

## FastAPI Endpoints for HTMX Partials

Biscuit works perfectly with HTMX for partial page updates when used with FastAPI. Endpoints can return HTML fragments generated by Biscuit views:

```python
@app.get("/comments/{post_id}")
view Comments(post_id: int, db: Session = Depends(get_db)):
    comments = db.query(Comment).filter(Comment.post_id == post_id).all()
    
    <div id="comments">
        for comment in comments:
            <div class="comment">
                <h4>{comment.author}</h4>
                <p>{comment.text}</p>
            </div>
    </div>

@app.post("/comments/{post_id}")
view AddComment(
    post_id: int,
    author: str = Form(...),
    text: str = Form(...),
    db: Session = Depends(get_db)
):
    # Add comment to database
    comment = Comment(post_id=post_id, author=author, text=text)
    db.add(comment)
    db.commit()
    
    # Return just the new comment markup for HTMX to insert
    <div class="comment">
        <h4>{comment.author}</h4>
        <p>{comment.text}</p>
    </div>
```

## Router Integration

Views work with FastAPI's `APIRouter` for modular applications:

```python
from fastapi import APIRouter, Depends

router = APIRouter(prefix="/admin")

@router.get("/dashboard")
view AdminDashboard(
    request: Request,
    analytics: AnalyticsData = Depends(get_analytics)
):
    <div class="admin-dashboard">
        <h1>Admin Dashboard</h1>
        <p>Welcome, {request.session["username"]}</p>
        
        <div class="stats">
            <div class="stat-card">
                <h3>Total Users</h3>
                <p class="number">{analytics.total_users}</p>
            </div>
            <div class="stat-card">
                <h3>Active Today</h3>
                <p class="number">{analytics.active_today}</p>
            </div>
        </div>
    </div>
```

## Automatic HTMLResponse

All Biscuit views automatically return `HTMLResponse` without needing to specify `response_class`:

```python
# This:
@app.get("/about")
view About():
    <div>About page content</div>

# Is equivalent to:
@app.get("/about", response_class=HTMLResponse)
def about():
    return HTMLResponse("<div>About page content</div>")
```

This simplifies route declarations while maintaining compatibility with FastAPI's response system.

# Special Features

- **Decorators**: Views can be decorated with Python decorators, including FastAPI route decorators
- **Typed parameters**: Full support for Python type annotations
- **Default parameters**: Support for default parameter values
- **Variable parameters**: Support for *args and **kwargs
- **Conditional rendering**: Using Python if/elif/else
- **List rendering**: Using Python for loops
- **Children and slots**: For flexible view composition
- **Dynamic attributes**: Parameters can use any Python expression
- **Intelligent tag resolution**: Automatically determines if a tag is an HTML element or a view based on naming convention and scope
- **Multiple root elements**: Views can contain multiple top-level HTML elements, unlike traditional components
- **HTMX integration**: First-class support for HTMX attributes for dynamic interactions
- **FastAPI integration**: Views can be used directly as route handlers