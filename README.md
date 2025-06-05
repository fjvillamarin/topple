# Sylfie: Python-Powered Web Templating Language

Sylfie is a modern templating language that seamlessly blends Python's power with HTML markup. It compiles `.psx` (Python Sylfie eXtension) files into pure Python code, enabling developers to build dynamic web UIs using familiar Python constructs.

## Key Features

- **Full Python Integration**: Use all Python features directly in templates - variables, loops, conditionals, functions, classes, comprehensions, and more
- **Component-Based Architecture**: Build reusable UI components called "views" with parameters and composition support
- **Smart Caching**: Built-in render caching prevents unnecessary re-rendering for better performance
- **Type Safety**: Full support for Python type hints and annotations
- **Modern Web Framework Support**: Designed for seamless integration with FastAPI, Flask, and other Python web frameworks
- **HTMX Ready**: First-class support for building interactive UIs with HTMX attributes
- **Fragment Support**: Render multiple root elements without wrapper divs

## Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/sylfie.git
cd sylfie

# Build the Sylfie compiler (requires Go 1.23+)
make build

# Copy the binary to your PATH
cp bin/sylfie /usr/local/bin/
# Or add bin/ to your PATH
export PATH="$PATH:$(pwd)/bin"
```

## Quick Start

### 1. Create a PSX View

Create a file named `hello.psx`:

```python
view HelloWorld(name: str = "World"):
    <div class="greeting">
        <h1>Hello, {name}!</h1>
        if len(name) > 10:
            <p>That's quite a long name!</p>
        else:
            <p>Nice to meet you.</p>
        
        <ul>
            for i in range(3):
                <li>Item {i + 1}</li>
        </ul>
    </div>
```

### 2. Compile to Python

```bash
sylfie compile hello.psx
```

This generates `hello.py` with a pure Python class that can be imported and used in any Python application.

### 3. Use in Your Application

```python
from hello import HelloWorld

# Create an instance and render
view = HelloWorld(name="Alice")
html = view.render()
print(html)
```

### Runtime Dependencies

The generated Python code requires the `runtime.py` file from the Sylfie project. Copy it to your project or ensure it's in your Python path.

## CLI Usage

The Sylfie compiler provides several commands:

```bash
# Compile a single file
sylfie compile input.psx

# Compile with custom output
sylfie compile input.psx -o output.py

# Compile directory recursively
sylfie compile src/ -r

# Watch mode for development
sylfie watch src/

# Debug: show tokens
sylfie scan input.psx

# Debug: show AST
sylfie parse input.psx

# Debug: show with resolution info
sylfie parse input.psx -d
```

## Language Features

### Views

Views are the core building blocks in PSX:

```python
view UserCard(user: User, show_email: bool = False):
    <div class="user-card">
        <h3>{user.name}</h3>
        if show_email:
            <p>{user.email}</p>
    </div>
```

### Control Flow

Use Python's control flow directly:

```python
view TodoList(todos: list[Todo]):
    <ul>
        for todo in todos:
            <li class={"done" if todo.completed else ""}>
                {todo.title}
            </li>
        else:
            <li>No todos yet!</li>
    </ul>
```

### Composition and Slots

Compose views within other views using the slot system:

```python
view Page(title: str, *, children=None):
    <html>
        <head><title>{title}</title></head>
        <body>
            <Header />
            <main>
                {render_child(children)}
            </main>
            <Footer />
        </body>
    </html>

view HomePage():
    <Page title="Home">
        <h1>Welcome!</h1>
        <UserCard user={current_user} />
    </Page>
```

### Python Integration

Use any Python expression or statement:

```python
view DataTable(data: list[dict]):
    data_frame = pd.DataFrame(data)
    stats = data_frame.describe()
    
    <div class="data-table">
        <h3>Data Summary</h3>
        <pre>{stats.to_string()}</pre>
        
        <table>
            for row in data_frame.itertuples():
                <tr>
                    for value in row[1:]:
                        <td>{value}</td>
                </tr>
        </table>
    </div>
```

## FastAPI Integration

```python
from fastapi import FastAPI
from fastapi.responses import HTMLResponse
from my_views import UserProfile

app = FastAPI()

@app.get("/user/{user_id}", response_class=HTMLResponse)
async def get_user(user_id: int):
    user = await fetch_user(user_id)
    view = UserProfile(user=user)
    return view.render()
```

## HTMX Support

```python
view SearchBox():
    <div>
        <input 
            type="search" 
            name="q"
            hx-get="/search"
            hx-trigger="keyup changed delay:500ms"
            hx-target="#results"
        />
        <div id="results"></div>
    </div>
```

## Architecture

Sylfie follows a multi-stage compilation process:

1. **Lexical Analysis**: Tokenizes `.psx` files into a stream of tokens
2. **Parsing**: Builds an Abstract Syntax Tree (AST) from tokens  
3. **Resolution**: Analyzes variable scopes and references
4. **Transformation**: Converts PSX-specific constructs (views, HTML elements) to Python AST
5. **Code Generation**: Outputs formatted Python code

The compiler is written in Go for performance and reliability.

## Development

### Project Structure

```
sylfie/
├── cmd/                  # CLI commands (compile, watch, scan, parse)
├── compiler/            # Core compiler implementation
│   ├── ast/            # AST node definitions
│   ├── lexer/          # Tokenization
│   ├── parser/         # AST construction
│   ├── resolver/       # Scope analysis
│   ├── transformers/   # View to Python AST transformation
│   └── codegen/        # Python code generation
├── runtime.py          # Python runtime library
├── bin/               # Compiled binaries
└── .mise.toml         # Task runner configuration
```

### Building from Source

```bash
# Requirements: Go 1.23+, mise

# Install mise (if not already installed)
curl https://mise.run | sh

# Trust the project configuration
mise trust

# Build the compiler
mise run build

# Run unit tests
mise run test

# Run all tests including golden file tests
mise run test-all

# Run specific test categories
mise run test-golden-category CATEGORY=views

# Update golden files after changes
mise run test-golden-update
```

### Development Commands

```bash
# View all available tasks
mise run help

# Format Go code
mise run fmt

# Run examples
mise run run

# Start watch mode
mise run watch

# View available test categories
mise run test-golden-list
```

### Contributing

1. Fork the repository
2. Create a feature branch
3. Write tests for your changes
4. Ensure all tests pass
5. Submit a pull request

## Current Status

Sylfie is in active development. Core features are implemented and working:

- ✅ Full Python syntax support
- ✅ View compilation to Python classes
- ✅ HTML element generation
- ✅ Control flow (if/for/while/match)
- ✅ Expression interpolation
- ✅ Smart render caching
- ✅ Fragment support
- ✅ Comprehensive test suite

### Known Limitations

- Generated code requires the `runtime.py` file
- Limited error recovery during parsing
- No source maps for debugging

## License

[License information to be added]

## Documentation

For more detailed information:

- [CLAUDE.md](CLAUDE.md) - Development guide for AI assistants
- Example code in `compiler/testdata/` directory
- Test categories showcase various language features