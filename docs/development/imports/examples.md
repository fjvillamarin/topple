# Import System Examples and Use Cases

## Overview

This document provides concrete examples of import functionality, expected behavior, and edge cases for the Topple import system.

## Basic Import Scenarios

### Example 1: Importing a View Component

**File: `components/button.psx`**
```python
view Button(text: str, variant: str = "primary"):
    class_name = f"btn btn-{variant}"
    <button class={class_name}>{text}</button>
```

**File: `app.psx`**
```python
from components.button import Button

view App():
    <div>
        <h1>My App</h1>
        <Button text="Click Me" />
        <Button text="Submit" variant="success" />
    </div>
```

**Expected Generated Output: `app.py`**
```python
from topple.psx import BaseView, Element, el, escape, fragment
from components.button import Button

class App(BaseView):
    def __init__(self):
        super().__init__()

    def _render(self) -> Element:
        return el("div",
            el("h1", "My App"),
            Button(text="Click Me")._render(),
            Button(text="Submit", variant="success")._render(),
        )
```

---

### Example 2: Importing Utility Functions

**File: `utils/formatting.psx`**
```python
def format_price(amount: float) -> str:
    return f"${amount:.2f}"

def format_date(date: datetime) -> str:
    return date.strftime("%Y-%m-%d")
```

**File: `views/product.psx`**
```python
from datetime import datetime
from utils.formatting import format_price, format_date

view ProductCard(name: str, price: float, updated: datetime):
    <div class="product-card">
        <h3>{name}</h3>
        <p class="price">{format_price(price)}</p>
        <p class="date">Updated: {format_date(updated)}</p>
    </div>
```

**Expected Behavior**:
- `format_price` and `format_date` are resolved as imported symbols
- Functions are available in view scope
- Generated code includes proper imports

---

### Example 3: Relative Imports

**Project Structure**:
```
project/
├── views/
│   ├── components/
│   │   ├── header.psx
│   │   └── footer.psx
│   └── layouts/
│       └── main.psx
```

**File: `views/layouts/main.psx`**
```python
from ..components.header import Header
from ..components.footer import Footer

view MainLayout(title: str, children: list):
    <div class="layout">
        <Header title={title} />
        <main>
            {children}
        </main>
        <Footer />
    </div>
```

**Expected Behavior**:
- `..components.header` resolves to `views/components/header.psx`
- Two levels up with `..`
- Imports work relative to current file location

---

### Example 4: Package Imports

**Project Structure**:
```
project/
├── components/
│   ├── __init__.psx
│   ├── button.psx
│   └── card.psx
```

**File: `components/__init__.psx`**
```python
from .button import Button
from .card import Card

__all__ = ["Button", "Card"]
```

**File: `app.psx`**
```python
from components import Button, Card

view App():
    <div>
        <Button text="Hello" />
        <Card title="Welcome" />
    </div>
```

**Expected Behavior**:
- Importing from `components` loads `components/__init__.psx`
- Re-exported symbols are available
- Package acts as namespace

---

### Example 5: Wildcard Imports

**File: `theme/colors.psx`**
```python
PRIMARY = "#007bff"
SECONDARY = "#6c757d"
SUCCESS = "#28a745"
DANGER = "#dc3545"

def get_color(name: str) -> str:
    colors = {
        "primary": PRIMARY,
        "secondary": SECONDARY,
        "success": SUCCESS,
        "danger": DANGER,
    }
    return colors.get(name, "#000000")
```

**File: `components/styled.psx`**
```python
from theme.colors import *

view StyledButton(text: str, color: str):
    bg_color = get_color(color)
    <button style=f"background-color: {bg_color}">{text}</button>
```

**Expected Behavior**:
- All public symbols from `colors.psx` imported
- `PRIMARY`, `SECONDARY`, etc. available
- `get_color` function available
- Private symbols (starting with `_`) excluded

---

### Example 6: Import Aliases

**File: `app.psx`**
```python
from components.button import Button as Btn
from components.card import Card as CardComponent
from utils.formatting import format_price as price

view ProductList(products: list):
    <div>
        {[
            <CardComponent key={p.id}>
                <h3>{p.name}</h3>
                <p>{price(p.price)}</p>
                <Btn text="Buy Now" />
            </CardComponent>
            for p in products
        ]}
    </div>
```

**Expected Behavior**:
- Imported symbols bound to aliases
- Aliases used in generated code
- Original names not accessible

---

## Advanced Scenarios

### Example 7: Multi-Level Dependencies

**Project Structure**:
```
project/
├── models/
│   └── user.psx
├── services/
│   └── auth.psx
└── views/
    └── profile.psx
```

**File: `models/user.psx`**
```python
class User:
    def __init__(self, name: str, email: str):
        self.name = name
        self.email = email
```

**File: `services/auth.psx`**
```python
from models.user import User

def get_current_user() -> User:
    return User(name="John Doe", email="john@example.com")
```

**File: `views/profile.psx`**
```python
from services.auth import get_current_user

view ProfileView():
    user = get_current_user()
    <div class="profile">
        <h1>{user.name}</h1>
        <p>{user.email}</p>
    </div>
```

**Expected Compilation Order**:
1. `models/user.psx` (no dependencies)
2. `services/auth.psx` (depends on user)
3. `views/profile.psx` (depends on auth)

**Expected Behavior**:
- Dependency graph correctly ordered
- Each file compiled with access to symbols from dependencies
- Transitive dependencies handled (profile uses User indirectly)

---

### Example 8: View Composition Across Files

**File: `components/layout.psx`**
```python
view Container(title: str):
    <div class="container">
        <header>
            <h1>{title}</h1>
        </header>
        <slot name="content" />
    </div>
```

**File: `pages/home.psx`**
```python
from components.layout import Container

view HomePage():
    <Container title="Welcome">
        <slot name="content">
            <p>This is the home page</p>
            <button>Get Started</button>
        </slot>
    </Container>
```

**Expected Behavior**:
- Views can be composed across file boundaries
- Slot mechanism works with imported views
- Proper nesting and rendering

---

## Error Cases

### Error 1: Module Not Found

**File: `app.psx`**
```python
from components.button import Button  # components/button.psx doesn't exist
```

**Expected Error**:
```
Error: cannot resolve import 'components.button'
  in file: app.psx
  searched paths:
    - ./components/button.psx
    - ./components/button/__init__.psx
```

---

### Error 2: Symbol Not Found

**File: `components/button.psx`**
```python
view Button():
    <button>Click</button>
```

**File: `app.psx`**
```python
from components.button import Card  # Card doesn't exist in button.psx
```

**Expected Error**:
```
Error: cannot import 'Card' from 'components.button'
  in file: app.psx:1
  available symbols: Button
```

---

### Error 3: Circular Dependency

**File: `a.psx`**
```python
from b import ViewB

view ViewA():
    <ViewB />
```

**File: `b.psx`**
```python
from a import ViewA

view ViewB():
    <ViewA />
```

**Expected Error**:
```
Error: circular dependencies detected:
  Cycle 1:
    a.psx
     ↓ imports
    b.psx
     ↓ imports (circular)
    a.psx
```

---

### Error 4: Invalid Relative Import

**File: `views/app.psx`**
```python
from ...components import Button  # Too many levels up
```

**Expected Error**:
```
Error: invalid relative import in 'views/app.psx'
  attempted to go 3 levels up from 'views/'
  project root is at '.'
```

---

## Integration with Existing Features

### Example 9: Imports with View Parameters

**File: `components/input.psx`**
```python
view TextInput(name: str, placeholder: str = "", value: str = ""):
    <input
        type="text"
        name={name}
        placeholder={placeholder}
        value={value}
    />
```

**File: `forms/login.psx`**
```python
from components.input import TextInput

view LoginForm():
    <form>
        <TextInput name="username" placeholder="Enter username" />
        <TextInput name="password" placeholder="Enter password" />
        <button type="submit">Login</button>
    </form>
```

**Expected Behavior**:
- View parameters resolved correctly
- Default parameters work across files
- Type hints preserved

---

### Example 10: Imports with Global Variables

**File: `config/settings.psx`**
```python
DEBUG = True
API_BASE_URL = "https://api.example.com"
MAX_RESULTS = 100
```

**File: `services/api.psx`**
```python
from config.settings import API_BASE_URL, MAX_RESULTS

def fetch_data(endpoint: str):
    url = f"{API_BASE_URL}/{endpoint}?limit={MAX_RESULTS}"
    # ... fetch logic
    return url
```

**Expected Behavior**:
- Global variables imported as symbols
- Values available at module level
- Constants usable in imported file

---

## Testing Scenarios

### Test Case 1: Simple Import Chain

**Input Files**:

`testdata/input/imports/chain/base.psx`:
```python
view Base():
    <div>Base</div>
```

`testdata/input/imports/chain/middle.psx`:
```python
from base import Base

view Middle():
    <div>
        <Base />
    </div>
```

`testdata/input/imports/chain/top.psx`:
```python
from middle import Middle

view Top():
    <div>
        <Middle />
    </div>
```

**Expected Compilation Order**:
1. `base.psx`
2. `middle.psx`
3. `top.psx`

**Expected Output Files**:

`testdata/expected/imports/chain/base.py`:
```python
from topple.psx import BaseView, Element, el, escape, fragment

class Base(BaseView):
    def __init__(self):
        super().__init__()

    def _render(self) -> Element:
        return el("div", "Base")
```

`testdata/expected/imports/chain/middle.py`:
```python
from topple.psx import BaseView, Element, el, escape, fragment
from base import Base

class Middle(BaseView):
    def __init__(self):
        super().__init__()

    def _render(self) -> Element:
        return el("div", Base()._render())
```

`testdata/expected/imports/chain/top.py`:
```python
from topple.psx import BaseView, Element, el, escape, fragment
from middle import Middle

class Top(BaseView):
    def __init__(self):
        super().__init__()

    def _render(self) -> Element:
        return el("div", Middle()._render())
```

---

### Test Case 2: Multiple Imports from Same Module

**Input**: `testdata/input/imports/multiple/utils.psx`
```python
def add(a: int, b: int) -> int:
    return a + b

def multiply(a: int, b: int) -> int:
    return a * b

def divide(a: int, b: int) -> float:
    return a / b
```

**Input**: `testdata/input/imports/multiple/app.psx`
```python
from utils import add, multiply

view Calculator():
    result = add(5, multiply(2, 3))
    <div>Result: {result}</div>
```

**Expected**: Both functions imported and usable

---

### Test Case 3: Relative Import Sibling

**Input**: `testdata/input/imports/relative/sibling/a.psx`
```python
view ViewA():
    <div>A</div>
```

**Input**: `testdata/input/imports/relative/sibling/b.psx`
```python
from .a import ViewA

view ViewB():
    <div>
        <ViewA />
    </div>
```

**Expected**: `.a` resolves to `sibling/a.psx` relative to `b.psx`

---

## Performance Test Cases

### Test Case: Large Project

**Structure**:
```
project/
├── components/ (50 files)
├── layouts/ (10 files)
├── pages/ (100 files)
└── utils/ (20 files)
```

**Expected Performance**:
- Parse all files: < 2s
- Build dependency graph: < 500ms
- Compile all files: < 5s
- Total: < 8s for 180 files

**Memory Usage**:
- Peak memory: < 200MB
- Per-file average: < 1MB

---

## Edge Cases

### Edge Case 1: Import Same Module with Different Aliases

**File**: `app.psx`
```python
from utils import format_price as price
from utils import format_price as fmt_price

view App():
    val1 = price(10.5)
    val2 = fmt_price(20.0)
    <div>{val1}, {val2}</div>
```

**Expected**: Both aliases work, bound to same symbol

---

### Edge Case 2: Shadowing Imported Names

**File**: `app.psx`
```python
from utils import helper

def helper():  # Shadows imported helper
    return "local"

view App():
    result = helper()  # Uses local, not imported
    <div>{result}</div>
```

**Expected**: Warning about shadowing, local definition takes precedence

---

### Edge Case 3: Import Inside Function (Not Supported)

**File**: `app.psx`
```python
view App():
    from utils import helper  # ERROR: imports must be top-level
    <div>{helper()}</div>
```

**Expected**: Compile error - imports only allowed at module level

---

### Edge Case 4: Star Import Conflict

**File**: `a.psx`
```python
VALUE = 10
```

**File**: `b.psx`
```python
VALUE = 20
```

**File**: `app.psx`
```python
from a import *
from b import *  # Shadows VALUE from a

view App():
    <div>{VALUE}</div>  # Should be 20
```

**Expected**: Later import wins, possible warning about shadowing

---

## Migration Examples

### Before (Single File)

**File**: `app.psx`
```python
view Button(text: str):
    <button>{text}</button>

view Card(title: str):
    <div class="card">
        <h3>{title}</h3>
    </div>

view App():
    <div>
        <Card title="Hello" />
        <Button text="Click" />
    </div>
```

### After (Multi-File with Imports)

**File**: `components/button.psx`
```python
view Button(text: str):
    <button>{text}</button>
```

**File**: `components/card.psx`
```python
view Card(title: str):
    <div class="card">
        <h3>{title}</h3>
    </div>
```

**File**: `app.psx`
```python
from components.button import Button
from components.card import Card

view App():
    <div>
        <Card title="Hello" />
        <Button text="Click" />
    </div>
```

**Benefits**:
- Better code organization
- Reusable components
- Easier testing
- Clearer dependencies

---

*These examples should guide implementation and testing of the import system.*
