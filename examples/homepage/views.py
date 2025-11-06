from topple.psx import BaseView, Element, el, escape, fragment
class LandingPage(BaseView):
    def __init__(self):
        super().__init__()

    def _render(self) -> Element:
        _root_children_1000 = []
        features = [{"title": "Type Safe", "desc": "Full Python type hints and IDE support", "icon": "🔒"}, {"title": "Server Rendered", "desc": "Fast SSR using Python runtime", "icon": "⚡"}, {"title": "JSX Familiar", "desc": "Write components like React but in Python", "icon": "⚛️"}]
        examples = [{"label": "For Loops", "code": "for item in items:\\n    <li>{item}</li>"}, {"label": "Conditionals", "code": "if show:\\n    <div>Visible</div>"}, {"label": "Nested", "code": "for row in rows:\\n    <tr>\\n        for cell in row:\\n            <td>{cell}</td>\\n    </tr>"}]
        _html_children_2000 = []
        _html_children_2000.append(el("head", [el("meta", "", {"charset": "UTF-8"}), el("meta", "", {"name": "viewport", "content": "width=device-width, initial-scale=1.0"}), el("title", "Topple - PSX"), el("script", "", {"src": "https://cdn.tailwindcss.com"})]))
        _body_children_3000 = []
        _div_children_4000 = []
        _div_children_4000.append(el("div", [el("h1", "Topple", {"class": "text-8xl font-bold text-transparent bg-clip-text bg-gradient-to-r from-purple-400 to-pink-600 mb-6"}), el("p", "PSX: HTML-like syntax in Python", {"class": "text-3xl text-purple-300 mb-4"}), el("p", "Write components using familiar JSX syntax, render on the server using Python", {"class": "text-xl text-slate-400 mb-8"}), el("a", "Get Started", {"href": "https://github.com/topple-lang/topple", "class": "inline-block px-8 py-4 bg-gradient-to-r from-purple-500 to-pink-500 text-white font-semibold rounded-lg shadow-lg hover:shadow-xl transform hover:scale-105 transition-all"})], {"class": "text-center mb-20"}))
        _div_children_4000.append(el("div", el("div", [el("div", [el("h3", "example.psx", {"class": "text-purple-400 font-mono text-sm mb-3"}), el("div", [el("div", "view Greeting(name: str):"), el("div", "&lt;div class=\"card\"&gt;", {"class": "ml-4"}), el("div", "&lt;h1&gt;Hello, name&lt;/h1&gt;", {"class": "ml-8"}), el("div", "&lt;/div&gt;", {"class": "ml-4"})], {"class": "bg-slate-900 rounded-lg p-4 font-mono text-sm text-slate-300"})]), el("div", [el("h3", "output.py", {"class": "text-pink-400 font-mono text-sm mb-3"}), el("div", [el("div", "class Greeting(BaseView):"), el("div", "def __init__(self, name: str):", {"class": "ml-4"}), el("div", "self.name = name", {"class": "ml-8"}), el("div", "def _render(self):", {"class": "ml-4"}), el("div", "return el(\"div\", ...)", {"class": "ml-8"})], {"class": "bg-slate-900 rounded-lg p-4 font-mono text-sm text-slate-300"})])], {"class": "grid md:grid-cols-2 gap-8"}), {"class": "bg-slate-800/50 backdrop-blur rounded-xl border border-purple-500/30 p-8 mb-20"}))
        _div_children_5000 = []
        for feature in features:
            _div_children_5000.append(el("div", [el("div", escape(feature["icon"]), {"class": "text-4xl mb-3"}), el("h3", escape(feature["title"]), {"class": "text-xl font-semibold text-purple-300 mb-2"}), el("p", escape(feature["desc"]), {"class": "text-slate-400"})], {"class": "bg-slate-800/30 border border-purple-500/20 rounded-lg p-6 hover:border-purple-500/40 transition-all"}))
        _div_children_4000.append(el("div", _div_children_5000, {"class": "grid md:grid-cols-3 gap-6 mb-20"}))
        _div_children_6000 = []
        _div_children_6000.append(el("h2", "Works with Compound Statements", {"class": "text-3xl font-bold text-purple-300 mb-6 text-center"}))
        _div_children_7000 = []
        for example in examples:
            _div_children_7000.append(el("div", [el("h4", escape(example["label"]), {"class": "text-purple-400 font-mono text-sm mb-3"}), el("pre", escape(example["code"]), {"class": "text-slate-300 text-xs font-mono whitespace-pre"})], {"class": "bg-slate-900 rounded-lg p-4"}))
        _div_children_6000.append(el("div", _div_children_7000, {"class": "grid md:grid-cols-3 gap-6"}))
        _div_children_4000.append(el("div", _div_children_6000, {"class": "bg-slate-800/50 backdrop-blur rounded-xl border border-purple-500/30 p-8"}))
        _body_children_3000.append(el("div", _div_children_4000, {"class": "container mx-auto px-4 max-w-6xl"}))
        _html_children_2000.append(el("body", _body_children_3000, {"class": "bg-gradient-to-br from-slate-900 via-purple-900 to-slate-900 min-h-screen py-16"}))
        _root_children_1000.append(el("html", _html_children_2000, {"lang": "en"}))
        return fragment(_root_children_1000)

