from topple.psx import BaseView, Element, el, escape, fragment
class EmptyState(BaseView):
    def __init__(self):
        super().__init__()

    def _render(self) -> Element:
        return el("div", [el("p", "No todos yet!", {"class": "text-lg"}), el("p", "Add one above to get started", {"class": "text-sm mt-2"})], {"class": "text-center py-12 text-slate-400"})

class TodoItem(BaseView):
    def __init__(self, todo: dict):
        super().__init__()
        self.todo = todo

    def _render(self) -> Element:
        return el("li", [el("span", escape(self.todo.get("text", "")), {"class": "text-slate-200"}), el("button", "Delete", {"hx-delete": escape(f"/todos/{self.todo.get("id", "")}"), "hx-target": "#todo-list", "hx-swap": "innerHTML", "class": "px-4 py-2 bg-red-500/20 text-red-400 border border-red-500/30 rounded hover:bg-red-500/30 transition-all"})], {"class": "flex items-center justify-between bg-slate-900/50 border border-purple-500/20 rounded-lg p-4 hover:border-purple-500/40 transition-all"})

class PageHeader(BaseView):
    def __init__(self, title: str, subtitle: str, description: str):
        super().__init__()
        self.title = title
        self.subtitle = subtitle
        self.description = description

    def _render(self) -> Element:
        return el("div", [el("h1", escape(self.title), {"class": "text-6xl font-bold text-transparent bg-clip-text bg-gradient-to-r from-purple-400 to-pink-600 mb-4"}), el("p", escape(self.subtitle), {"class": "text-xl text-purple-300"}), el("p", escape(self.description), {"class": "text-sm text-slate-400 mt-2"})], {"class": "text-center mb-12"})

class TodoForm(BaseView):
    def __init__(self):
        super().__init__()

    def _render(self) -> Element:
        return el("form", el("div", [el("input", "", {"type": "text", "name": "todo", "placeholder": "What needs to be done?", "required": "required", "class": "flex-1 px-4 py-3 bg-slate-900 border border-purple-500/30 rounded-lg text-slate-200 placeholder-slate-500 focus:outline-none focus:border-purple-500"}), el("button", "Add", {"type": "submit", "class": "px-6 py-3 bg-gradient-to-r from-purple-500 to-pink-500 text-white font-semibold rounded-lg hover:shadow-lg transition-all"})], {"class": "flex gap-3"}), {"hx-post": "/todos", "hx-target": "#todo-list", "hx-swap": "innerHTML", "class": "mb-8"})

class TodoList(BaseView):
    def __init__(self, todos: list):
        super().__init__()
        self.todos = todos

    def _render(self) -> Element:
        _root_children_3000 = []
        _div_children_4000 = []
        if len(self.todos) == 0:
            _div_children_4000.append(EmptyState())
        else:
            _ul_children_5000 = []
            for self.todo in self.todos:
                _ul_children_5000.append(TodoItem(todo=self.todo))
            _div_children_4000.append(el("ul", _ul_children_5000, {"class": "space-y-3"}))
        _root_children_3000.append(el("div", _div_children_4000))
        return fragment(_root_children_3000)

class TodoPage(BaseView):
    def __init__(self, todos: list):
        super().__init__()
        self.todos = todos

    def _render(self) -> Element:
        return el("html", [el("head", [el("meta", "", {"charset": "UTF-8"}), el("meta", "", {"name": "viewport", "content": "width=device-width, initial-scale=1.0"}), el("title", "PSX + HTMX Todo List"), el("script", "", {"src": "https://unpkg.com/htmx.org@1.9.10"}), el("script", "", {"src": "https://cdn.tailwindcss.com"})]), el("body", el("div", [PageHeader(title="PSX + HTMX", subtitle="Todo List Demo", description="Server-side rendering with zero JavaScript"), el("div", [TodoForm(), el("div", TodoList(todos=self.todos), {"id": "todo-list"})], {"class": "bg-slate-800/50 backdrop-blur rounded-xl border border-purple-500/30 p-8"}), el("div", [el("p", "This example demonstrates PSX server-side rendering with HTMX interactivity"), el("p", "All logic runs on the server - no JavaScript framework needed!", {"class": "mt-1"})], {"class": "mt-8 text-center text-slate-500 text-sm"})], {"class": "container mx-auto px-4 max-w-2xl"}), {"class": "bg-gradient-to-br from-slate-900 via-purple-900 to-slate-900 min-h-screen py-16"})], {"lang": "en"})

