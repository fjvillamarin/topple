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

### View Composition

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

#### Multiple Root Elements

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

#### Slots System

Biscuit provides a Vue.js-inspired slot system for flexible content distribution:

##### Basic Default Slot

The simplest form is the default (unnamed) slot:

```python
# Defining a view with a default slot
view Card():
    <div class="card">
        <slot />
    </div>

# Using the view with default slot content
<Card>
    <p>This content goes in the default slot</p>
</Card>
```

##### Named Slots

For more complex layouts, named slots allow multiple content areas:

```python
# Defining a view with named slots
view PageLayout():
    <div class="layout">
        <header><slot name="header" /></header>
        <main><slot /></main>
        <footer><slot name="footer" /></footer>
    </div>

# Using named slots
<PageLayout>
    <h1 slot="header">Page Title</h1>
    
    <p>Main content goes in the default slot</p>
    
    <p slot="footer">Footer content</p>
</PageLayout>
```

##### Fallback Content

Slots can provide fallback content when none is supplied:

```python
# Defining fallback content in slots
view Alert(type="info"):
    <div class="alert alert-{type}">
        <div class="alert-icon">
            <slot name="icon">
                <DefaultIcon type={type} />
            </slot>
        </div>
        <div class="alert-content">
            <slot>
                <p>Default alert message</p>
            </slot>
        </div>
    </div>

# Usage with or without providing content
<Alert type="warning">
    <p>Custom warning message</p>
</Alert>

<Alert type="error" />  <!-- Will use default content -->
```

##### Conditional Slot Rendering

Conditionally render elements based on whether a slot has content:

```python
view Panel(title):
    <div class="panel">
        <h3>{title}</h3>
        <div class="panel-body">
            <slot />
        </div>
        
        if has_slot("footer"):
            <div class="panel-footer">
                <slot name="footer" />
            </div>
    </div>
```

## Compilation Process

Biscuit files (`.bsct`) are compiled to Python code through:
1. Parsing the Biscuit syntax tree using tree-sitter
2. Transforming views into Python classes
3. Creating render methods that output HTML strings
4. Handling interpolation via Python f-strings
5. Resolving view references based on imports and scope

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

## HTMX Integration

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

#### HTMX Endpoint Views

You can create views that serve as HTMX endpoints, returning only the HTML fragments needed:

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

#### Dynamic Loading with HTMX

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

#### Form Processing with HTMX

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

#### Form Validation Response

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

#### Infinite Scroll with HTMX

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

#### Active Search with HTMX

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

#### Tabs with HTMX

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

## Error Management

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

## FastAPI Integration

Biscuit seamlessly integrates with FastAPI to create HTML responses:

### Basic Route Handlers

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

### Path Parameters

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

### Query Parameters

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

### Form Data

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

### Request Body

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

### HTMX Integration

Biscuit works perfectly with HTMX for partial page updates:

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

### Router Integration

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

### Automatic HTMLResponse

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

## Special Features

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