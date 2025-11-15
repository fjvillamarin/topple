from topple.psx import BaseView, Element, el, escape, fragment
class EmptyState(BaseView):
    def __init__(self):
        super().__init__()

    def _render(self) -> Element:
        _root_children_1000 = []
        _div_children_2000 = []
        _div_children_2000.append(el("p", "No todos yet!", {"class": "text-lg"}))
        _div_children_2000.append(el("p", "Add one above to get started", {"class": "text-sm mt-2"}))
        _root_children_1000.append(el("div", _div_children_2000, {"class": "text-center py-12 text-slate-400"}))
        return fragment(_root_children_1000)

class TodoItem(BaseView):
    def __init__(self, todo: dict):
        super().__init__()
        self.todo = todo

    def _render(self) -> Element:
        _root_children_3000 = []
        _li_children_4000 = []
        _li_children_4000.append(el("span", escape(self.todo.get("text", "")), {"class": "text-slate-200"}))
        _li_children_4000.append(el("button", "Delete", {"hx-delete": escape(f"/todos/{self.todo.get("id", "")}"), "hx-target": "#todo-list", "hx-swap": "innerHTML", "class": "px-4 py-2 bg-red-500/20 text-red-400 border border-red-500/30 rounded hover:bg-red-500/30 transition-all"}))
        _root_children_3000.append(el("li", _li_children_4000, {"class": "flex items-center justify-between bg-slate-900/50 border border-purple-500/20 rounded-lg p-4 hover:border-purple-500/40 transition-all"}))
        return fragment(_root_children_3000)

class PageHeader(BaseView):
    def __init__(self, title: str, subtitle: str, description: str):
        super().__init__()
        self.title = title
        self.subtitle = subtitle
        self.description = description

    def _render(self) -> Element:
        _root_children_5000 = []
        _div_children_6000 = []
        _div_children_6000.append(el("h1", escape(self.title), {"class": "text-6xl font-bold text-transparent bg-clip-text bg-gradient-to-r from-purple-400 to-pink-600 mb-4"}))
        _div_children_6000.append(el("p", escape(self.subtitle), {"class": "text-xl text-purple-300"}))
        _div_children_6000.append(el("p", escape(self.description), {"class": "text-sm text-slate-400 mt-2"}))
        _root_children_5000.append(el("div", _div_children_6000, {"class": "text-center mb-12"}))
        return fragment(_root_children_5000)

class TodoForm(BaseView):
    def __init__(self):
        super().__init__()

    def _render(self) -> Element:
        _root_children_7000 = []
        _form_children_8000 = []
        _div_children_9000 = []
        _div_children_9000.append(el("input", "", {"type": "text", "name": "todo", "placeholder": "What needs to be done?", "required": "required", "class": "flex-1 px-4 py-3 bg-slate-900 border border-purple-500/30 rounded-lg text-slate-200 placeholder-slate-500 focus:outline-none focus:border-purple-500"}))
        _div_children_9000.append(el("button", "Add", {"type": "submit", "class": "px-6 py-3 bg-gradient-to-r from-purple-500 to-pink-500 text-white font-semibold rounded-lg hover:shadow-lg transition-all"}))
        _form_children_8000.append(el("div", _div_children_9000, {"class": "flex gap-3"}))
        _root_children_7000.append(el("form", _form_children_8000, {"hx-post": "/todos", "hx-target": "#todo-list", "hx-swap": "innerHTML", "class": "mb-8"}))
        return fragment(_root_children_7000)

class TodoList(BaseView):
    def __init__(self, todos: list):
        super().__init__()
        self.todos = todos

    def _render(self) -> Element:
        _root_children_10000 = []
        _div_children_11000 = []
        if len(self.todos) == 0:
            _div_children_11000.append(EmptyState())
        else:
            _ul_children_12000 = []
            for todo in self.todos:
                _ul_children_12000.append(TodoItem(todo=todo))
            _div_children_11000.append(el("ul", _ul_children_12000, {"class": "space-y-3"}))
        _root_children_10000.append(el("div", _div_children_11000))
        return fragment(_root_children_10000)

class TodoPage(BaseView):
    def __init__(self, todos: list):
        super().__init__()
        self.todos = todos

    def _render(self) -> Element:
        _root_children_13000 = []
        _html_children_14000 = []
        _head_children_15000 = []
        _head_children_15000.append(el("meta", "", {"charset": "UTF-8"}))
        _head_children_15000.append(el("meta", "", {"name": "viewport", "content": "width=device-width, initial-scale=1.0"}))
        _head_children_15000.append(el("title", "PSX + HTMX Todo List"))
        _head_children_15000.append(el("script", "", {"src": "https://unpkg.com/htmx.org@2.0.6"}))
        _head_children_15000.append(el("script", "", {"src": "https://cdn.tailwindcss.com@4.1.16"}))
        _html_children_14000.append(el("head", _head_children_15000))
        _body_children_16000 = []
        _div_children_17000 = []
        _div_children_17000.append(PageHeader(title="PSX + HTMX", subtitle="Todo List Demo", description="Server-side rendering with zero JavaScript"))
        _div_children_18000 = []
        _div_children_18000.append(TodoForm())
        _div_children_18000.append(el("div", TodoList(todos=self.todos), {"id": "todo-list"}))
        _div_children_17000.append(el("div", _div_children_18000, {"class": "bg-slate-800/50 backdrop-blur rounded-xl border border-purple-500/30 p-8"}))
        _div_children_19000 = []
        _div_children_19000.append(el("p", "This example demonstrates PSX server-side rendering with HTMX interactivity"))
        _div_children_19000.append(el("p", "All logic runs on the server - no JavaScript framework needed!", {"class": "mt-1"}))
        _div_children_17000.append(el("div", _div_children_19000, {"class": "mt-8 text-center text-slate-500 text-sm"}))
        _body_children_16000.append(el("div", _div_children_17000, {"class": "container mx-auto px-4 max-w-2xl"}))
        _html_children_14000.append(el("body", _body_children_16000, {"class": "bg-gradient-to-br from-slate-900 via-purple-900 to-slate-900 min-h-screen py-16"}))
        _root_children_13000.append(el("html", _html_children_14000, {"lang": "en"}))
        return fragment(_root_children_13000)

