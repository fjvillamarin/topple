# Topple Runtime System

The Topple runtime (`topple/psx.py`) provides the foundation for executing compiled PSX views. It includes classes and functions for HTML generation, automatic escaping, and view composition.

## Runtime Architecture

The runtime is designed with several key principles:

1. **Security First**: All dynamic content is automatically escaped
2. **Performance**: Multi-level caching prevents redundant rendering
3. **Simplicity**: Minimal API surface for generated code
4. **Type Safety**: Full type hints for better IDE support

## Core Components

### BaseView Class

The abstract base class for all compiled views:

```python
class BaseView(ABC):
    """Abstract base class for any PSX view rendered on the server."""
    
    def __init__(self):
        """Initialize cache attributes."""
        self._render_cache: Optional[Union[Element, str]] = None
        self._html_cache: Optional[str] = None
    
    @abstractmethod
    def _render(self) -> Union[Element, str]:
        """Return either an Element instance or a plain string."""
        ...
    
    def render(self) -> str:
        """Public interface that ensures string output with caching."""
        if self._html_cache is not None:
            return self._html_cache
        
        result = self._get_rendered()
        self._html_cache = str(result)
        return self._html_cache
```

Key features:
- **Two-level caching**: Both the render result and final HTML are cached
- **Protected `_render()` method**: Implemented by generated code
- **Public `render()` method**: Used by application code

### Element Class

Represents HTML elements with automatic escaping:

```python
class Element:
    """In-memory representation of an HTML element."""
    
    def __init__(self, tag: str, children: ..., attrs: Dict[str, Any] = None, 
                 self_close: bool = False):
        self.tag = tag
        self.children = normalize_to_list(children)
        self.attrs = attrs or {}
        self.self_close = self_close
        self._html_cache: Optional[str] = None
```

Features:
- **Flexible children**: Can be strings, Elements, BaseViews, or lists
- **Attribute handling**: Supports boolean attributes and dynamic values
- **Self-closing tags**: Proper handling of void elements
- **Caching**: Each element caches its rendered HTML

### FragmentElement Class

A special element that renders children without a wrapper:

```python
class FragmentElement(Element):
    """Renders children without any wrapper tag (like React fragments)."""
    
    def __str__(self) -> str:
        # Concatenates children without wrapper tags
        return "".join(render_each_child(self.children))
```

## Core Functions

### escape()

Security-critical function for HTML escaping:

```python
def escape(raw: Any) -> str:
    """Convert raw data into a safely-escaped string for HTML output."""
    if raw is None:
        return ""
    if isinstance(raw, str):
        return html.escape(raw, quote=True)  # Escapes &, <, >, ", '
    return str(raw)
```

### el()

Factory function for creating elements:

```python
def el(tag: str, content: ... = "", attrs: Dict[str, Any] = None, 
       self_close: bool = False) -> Element:
    """Create an Element for the given tag, children, and attributes."""
    return Element(tag, content, attrs, self_close)
```

### render_child()

Normalizes nested content for proper rendering:

```python
def render_child(child: Union[BaseView, Element, str, ...]) -> Union[Element, str]:
    """Normalize content to Element or string for later escaping."""
    if child is None:
        return ""
    if isinstance(child, BaseView):
        return child._get_rendered()  # Uses caching
    if isinstance(child, Element):
        return child
    return str(child)  # Will be escaped when rendered
```

### fragment()

Creates a fragment element:

```python
def fragment(children: List[Union[BaseView, Element, str]]) -> FragmentElement:
    """Create a FragmentElement that concatenates children without wrapper."""
    return FragmentElement(children)
```

## Compilation Examples

### Basic View

**Input PSX:**
```python
view HelloWorld():
    <div>Hello World</div>
```

**Generated Python:**
```python
from topple.psx import BaseView, el, escape, Element, FragmentElement, fragment, render_child

class HelloWorld(BaseView):
    def __init__(self):
        super().__init__()
    
    def _render(self) -> Element:
        return el("div", "Hello World")
```

### View with Parameters and Escaping

**Input PSX:**
```python
view Greeting(name: str, age: int = 25):
    <div>
        <h1>Hello, {name}!</h1>
        <p>You are {age} years old.</p>
    </div>
```

**Generated Python:**
```python
class Greeting(BaseView):
    def __init__(self, name: str, age: int = 25):
        super().__init__()
        self.name = name
        self.age = age
    
    def _render(self) -> Element:
        return el("div", [
            el("h1", f"Hello, {escape(self.name)}!"),
            el("p", f"You are {escape(self.age)} years old.")
        ])
```

## Security Features

### Automatic HTML Escaping

All dynamic content is automatically escaped:

```python
# User input
user_input = "<script>alert('XSS')</script>"

# In view
<div>{user_input}</div>

# Generated code
el("div", escape(user_input))

# Rendered HTML
<div>&lt;script&gt;alert('XSS')&lt;/script&gt;</div>
```

### Attribute Escaping

Attributes are properly quoted and escaped:

```python
# Dangerous attribute
onclick_value = 'alert("XSS"); return false;'

# In view
<button onclick={onclick_value}>Click</button>

# Rendered HTML
<button onclick="alert(&quot;XSS&quot;); return false;">Click</button>
```

### Safe Boolean Attributes

Boolean attributes are handled correctly:

```python
# In view
<input type="checkbox" checked={is_checked} disabled={is_disabled} />

# When is_checked=True, is_disabled=False
<input type="checkbox" checked />

# When is_checked=False, is_disabled=True
<input type="checkbox" disabled />
```

## Performance Optimizations

### Multi-Level Caching

1. **BaseView level**: `_render()` result cached in `_render_cache`
2. **BaseView level**: Final HTML cached in `_html_cache`
3. **Element level**: Rendered HTML cached in `_html_cache`
4. **FragmentElement level**: Concatenated HTML cached

This prevents re-rendering when the same view/element is accessed multiple times.

### Efficient String Building

The runtime uses list concatenation for efficiency:

```python
# Instead of string concatenation
html = ""
for item in items:
    html += f"<li>{escape(item)}</li>"  # Inefficient

# The runtime uses
parts = []
for item in items:
    parts.append(el("li", escape(item)))
return "".join(parts)  # More efficient
```

## Advanced Features

### Conditional Rendering

```python
# In PSX
if user.is_authenticated:
    <div class="welcome">Welcome, {user.name}!</div>
else:
    <div class="login">Please log in</div>

# Generated
if self.user.is_authenticated:
    _view_children.append(
        el("div", f"Welcome, {escape(self.user.name)}!", {"class": "welcome"})
    )
else:
    _view_children.append(
        el("div", "Please log in", {"class": "login"})
    )
```

### Loop Rendering

```python
# In PSX
for item in items:
    <li>{item.name}: ${item.price}</li>

# Generated
for item in self.items:
    _view_children.append(
        el("li", f"{escape(item.name)}: ${escape(item.price)}")
    )
```

### HTMX Support

The runtime preserves HTMX attributes:

```python
# In PSX
<button 
    hx-get="/api/data"
    hx-target="#content"
    hx-swap="innerHTML"
>
    Load Data
</button>

# Generated
el("button", "Load Data", {
    "hx-get": "/api/data",
    "hx-target": "#content",
    "hx-swap": "innerHTML"
})
```

## Error Handling

The runtime provides clear error messages:

- **Missing required parameters**: TypeError with parameter name
- **Invalid child types**: Clear error about unsupported type
- **Rendering errors**: Stack trace points to view source

## Custom Runtime Extensions

You can extend the runtime by:

1. **Subclassing BaseView**:
```python
class CustomBaseView(BaseView):
    def __init__(self, request=None):
        super().__init__()
        self.request = request
    
    def url_for(self, name: str, **params) -> str:
        # Custom URL generation
        pass
```

2. **Adding helper functions**:
```python
def class_names(*classes: str, **conditionals: bool) -> str:
    """Build class string from conditionals."""
    result = list(classes)
    for cls, condition in conditionals.items():
        if condition:
            result.append(cls)
    return " ".join(result)
```

3. **Custom element types**:
```python
class LazyElement(Element):
    """Element that renders on demand."""
    def __init__(self, loader_func):
        self.loader_func = loader_func
        super().__init__("div", "")
    
    def __str__(self):
        if not self._html_cache:
            self.children = [self.loader_func()]
        return super().__str__()
```

## Best Practices

1. **Let the runtime handle escaping**: Never manually escape content
2. **Use type hints**: Helps catch errors and improves IDE support
3. **Leverage caching**: Views are cached, so avoid side effects in `_render()`
4. **Keep views pure**: Views should be deterministic functions of their inputs
5. **Use fragments for multiple roots**: When returning multiple elements without wrapper

## Framework Integration

### FastAPI
```python
from fastapi import FastAPI
from fastapi.responses import HTMLResponse

@app.get("/", response_class=HTMLResponse)
async def index():
    view = HomePage(user=current_user)
    return view.render()
```

### Flask
```python
from flask import Flask

@app.route("/")
def index():
    view = HomePage(user=current_user)
    return view.render()
```

### Django
```python
from django.http import HttpResponse

def index(request):
    view = HomePage(user=request.user)
    return HttpResponse(view.render())
```

## Debugging

The runtime preserves clear stack traces:

```
Traceback (most recent call last):
  File "app.py", line 10, in index
    return HomePage(user=None).render()
  File "runtime.py", line 273, in render
    result = self._get_rendered()
  File "views.py", line 8, in _render
    el("h1", f"Welcome, {escape(self.user.name)}")
AttributeError: 'NoneType' object has no attribute 'name'
```

## See Also

- [Language Grammar](grammar_psx.md) - PSX syntax reference
- [Architecture Guide](architecture.md) - Compiler internals
- [Examples](../examples/) - Example applications