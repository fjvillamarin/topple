from topple.psx import BaseView, Element, el, escape, fragment
class FeatureCard(BaseView):
    def __init__(self, icon: str, title: str, description: str):
        super().__init__()
        self.icon = icon
        self.title = title
        self.description = description

    def _render(self) -> Element:
        _root_children_1000 = []
        _div_children_2000 = []
        _div_children_2000.append(el("div", escape(self.icon), {"class": "text-4xl mb-3"}))
        _div_children_2000.append(el("h3", escape(self.title), {"class": "text-xl font-semibold text-purple-300 mb-2"}))
        _div_children_2000.append(el("p", escape(self.description), {"class": "text-slate-400"}))
        _root_children_1000.append(el("div", _div_children_2000, {"class": "bg-slate-800/30 border border-purple-500/20 rounded-lg p-6 hover:border-purple-500/40 transition-all"}))
        return fragment(_root_children_1000)

class ExampleCard(BaseView):
    def __init__(self, label: str, code: str):
        super().__init__()
        self.label = label
        self.code = code

    def _render(self) -> Element:
        _root_children_3000 = []
        _div_children_4000 = []
        _div_children_4000.append(el("h4", escape(self.label), {"class": "text-purple-400 font-mono text-sm mb-3"}))
        _div_children_4000.append(el("pre", escape(self.code), {"class": "text-slate-300 text-xs font-mono whitespace-pre"}))
        _root_children_3000.append(el("div", _div_children_4000, {"class": "bg-slate-900 rounded-lg p-4"}))
        return fragment(_root_children_3000)

class PageHeader(BaseView):
    def __init__(self, title: str, subtitle: str, description: str, cta_text: str, cta_url: str):
        super().__init__()
        self.title = title
        self.subtitle = subtitle
        self.description = description
        self.cta_text = cta_text
        self.cta_url = cta_url

    def _render(self) -> Element:
        _root_children_5000 = []
        _div_children_6000 = []
        _div_children_6000.append(el("h1", escape(self.title), {"class": "text-8xl font-bold text-transparent bg-clip-text bg-gradient-to-r from-purple-400 to-pink-600 mb-6"}))
        _div_children_6000.append(el("p", escape(self.subtitle), {"class": "text-3xl text-purple-300 mb-4"}))
        _div_children_6000.append(el("p", escape(self.description), {"class": "text-xl text-slate-400 mb-8"}))
        _div_children_6000.append(el("a", escape(self.cta_text), {"href": escape(self.cta_url), "class": "inline-block px-8 py-4 bg-gradient-to-r from-purple-500 to-pink-500 text-white font-semibold rounded-lg shadow-lg hover:shadow-xl transform hover:scale-105 transition-all"}))
        _root_children_5000.append(el("div", _div_children_6000, {"class": "text-center mb-20"}))
        return fragment(_root_children_5000)

class CodeComparison(BaseView):
    def __init__(self):
        super().__init__()

    def _render(self) -> Element:
        _root_children_7000 = []
        _div_children_8000 = []
        _div_children_9000 = []
        _div_children_10000 = []
        _div_children_10000.append(el("h3", "example.psx", {"class": "text-purple-400 font-mono text-sm mb-3"}))
        _div_children_11000 = []
        _div_children_11000.append(el("div", "view Greeting(name: str):"))
        _div_children_11000.append(el("div", "&lt;div class=\"card\"&gt;", {"class": "ml-4"}))
        _div_children_11000.append(el("div", "&lt;h1&gt;Hello, name&lt;/h1&gt;", {"class": "ml-8"}))
        _div_children_11000.append(el("div", "&lt;/div&gt;", {"class": "ml-4"}))
        _div_children_10000.append(el("div", _div_children_11000, {"class": "bg-slate-900 rounded-lg p-4 font-mono text-sm text-slate-300"}))
        _div_children_9000.append(el("div", _div_children_10000))
        _div_children_12000 = []
        _div_children_12000.append(el("h3", "output.py", {"class": "text-pink-400 font-mono text-sm mb-3"}))
        _div_children_13000 = []
        _div_children_13000.append(el("div", "class Greeting(BaseView):"))
        _div_children_13000.append(el("div", "def __init__(self, name: str):", {"class": "ml-4"}))
        _div_children_13000.append(el("div", "self.name = name", {"class": "ml-8"}))
        _div_children_13000.append(el("div", "def _render(self):", {"class": "ml-4"}))
        _div_children_13000.append(el("div", "return el(\"div\", ...)", {"class": "ml-8"}))
        _div_children_12000.append(el("div", _div_children_13000, {"class": "bg-slate-900 rounded-lg p-4 font-mono text-sm text-slate-300"}))
        _div_children_9000.append(el("div", _div_children_12000))
        _div_children_8000.append(el("div", _div_children_9000, {"class": "grid md:grid-cols-2 gap-8"}))
        _root_children_7000.append(el("div", _div_children_8000, {"class": "bg-slate-800/50 backdrop-blur rounded-xl border border-purple-500/30 p-8 mb-20"}))
        return fragment(_root_children_7000)

class FeaturesGrid(BaseView):
    def __init__(self, features: list):
        super().__init__()
        self.features = features

    def _render(self) -> Element:
        _root_children_14000 = []
        _div_children_15000 = []
        for feature in self.features:
            _div_children_15000.append(FeatureCard(icon=feature["icon"], title=feature["title"], description=feature["desc"]))
        _root_children_14000.append(el("div", _div_children_15000, {"class": "grid md:grid-cols-3 gap-6 mb-20"}))
        return fragment(_root_children_14000)

class ExamplesSection(BaseView):
    def __init__(self, examples: list):
        super().__init__()
        self.examples = examples

    def _render(self) -> Element:
        _root_children_16000 = []
        _div_children_17000 = []
        _div_children_17000.append(el("h2", "Works with Compound Statements", {"class": "text-3xl font-bold text-purple-300 mb-6 text-center"}))
        _div_children_18000 = []
        for example in self.examples:
            _div_children_18000.append(ExampleCard(label=example["label"], code=example["code"]))
        _div_children_17000.append(el("div", _div_children_18000, {"class": "grid md:grid-cols-3 gap-6"}))
        _root_children_16000.append(el("div", _div_children_17000, {"class": "bg-slate-800/50 backdrop-blur rounded-xl border border-purple-500/30 p-8"}))
        return fragment(_root_children_16000)

class LandingPage(BaseView):
    def __init__(self):
        super().__init__()

    def _render(self) -> Element:
        _root_children_19000 = []
        features = [{"title": "Type Safe", "desc": "Full Python type hints and IDE support", "icon": "🔒"}, {"title": "Server Rendered", "desc": "Fast SSR using Python runtime", "icon": "⚡"}, {"title": "JSX Familiar", "desc": "Write components like React but in Python", "icon": "⚛️"}]
        examples = [{"label": "For Loops", "code": "for item in items:\\n    <li>{item}</li>"}, {"label": "Conditionals", "code": "if show:\\n    <div>Visible</div>"}, {"label": "Nested", "code": "for row in rows:\\n    <tr>\\n        for cell in row:\\n            <td>{cell}</td>\\n    </tr>"}]
        _html_children_20000 = []
        _head_children_21000 = []
        _head_children_21000.append(el("meta", "", {"charset": "UTF-8"}))
        _head_children_21000.append(el("meta", "", {"name": "viewport", "content": "width=device-width, initial-scale=1.0"}))
        _head_children_21000.append(el("title", "Topple - PSX"))
        _head_children_21000.append(el("script", "", {"src": "https://cdn.tailwindcss.com"}))
        _html_children_20000.append(el("head", _head_children_21000))
        _body_children_22000 = []
        _div_children_23000 = []
        _div_children_23000.append(PageHeader(title="Topple", subtitle="PSX: HTML-like syntax in Python", description="Write components using familiar JSX syntax, render on the server using Python", cta_text="Get Started", cta_url="https://github.com/topple-lang/topple"))
        _div_children_23000.append(CodeComparison())
        _div_children_23000.append(FeaturesGrid(features=features))
        _div_children_23000.append(ExamplesSection(examples=examples))
        _body_children_22000.append(el("div", _div_children_23000, {"class": "container mx-auto px-4 max-w-6xl"}))
        _html_children_20000.append(el("body", _body_children_22000, {"class": "bg-gradient-to-br from-slate-900 via-purple-900 to-slate-900 min-h-screen py-16"}))
        _root_children_19000.append(el("html", _html_children_20000, {"lang": "en"}))
        return fragment(_root_children_19000)

