from topple.psx import BaseView, Element, el, escape, fragment, raw
class PageHeader(BaseView):
    def __init__(self, title: str, subtitle: str, description: str, cta_text: str, cta_url: str):
        super().__init__()
        self.title = title
        self.subtitle = subtitle
        self.description = description
        self.cta_text = cta_text
        self.cta_url = cta_url

    def _render(self) -> Element:
        _root_children_1000 = []
        _div_children_2000 = []
        _div_children_2000.append(el("h1", escape(self.title), {"class": "text-8xl font-bold text-transparent bg-clip-text bg-gradient-to-r from-purple-400 to-pink-600 mb-6"}))
        _div_children_2000.append(el("p", escape(self.subtitle), {"class": "text-3xl text-purple-300 mb-4"}))
        _div_children_2000.append(el("p", escape(self.description), {"class": "text-xl text-slate-400 mb-8"}))
        _div_children_2000.append(el("a", escape(self.cta_text), {"href": escape(self.cta_url), "class": "inline-block px-8 py-4 bg-gradient-to-r from-purple-500 to-pink-500 text-white font-semibold rounded-lg shadow-lg hover:shadow-xl transform hover:scale-105 transition-all"}))
        _root_children_1000.append(el("div", _div_children_2000, {"class": "text-center mb-20"}))
        return fragment(_root_children_1000)

class CodeComparison(BaseView):
    def __init__(self):
        super().__init__()

    def _render(self) -> Element:
        _root_children_3000 = []
        psx_example = "view Greeting(name: str):\n    <div class=\"card\">\n        <h1>Hello, {name}!</h1>\n    </div>"
        python_example = "class Greeting(BaseView):\n    def __init__(self, name: str):\n        self.name = name\n    def _render(self):\n        return el(\"div\", ...)"
        _div_children_4000 = []
        _div_children_5000 = []
        _div_children_6000 = []
        _div_children_6000.append(el("h3", "example.psx", {"class": "text-purple-400 font-mono text-sm mb-3"}))
        _div_children_6000.append(el("pre", el("code", escape(raw(psx_example))), {"class": "bg-slate-900 rounded-lg p-4 font-mono text-sm text-slate-300"}))
        _div_children_5000.append(el("div", _div_children_6000))
        _div_children_7000 = []
        _div_children_7000.append(el("h3", "output.py", {"class": "text-pink-400 font-mono text-sm mb-3"}))
        _div_children_7000.append(el("pre", el("code", escape(raw(python_example))), {"class": "bg-slate-900 rounded-lg p-4 font-mono text-sm text-slate-300"}))
        _div_children_5000.append(el("div", _div_children_7000))
        _div_children_4000.append(el("div", _div_children_5000, {"class": "grid md:grid-cols-2 gap-8"}))
        _root_children_3000.append(el("div", _div_children_4000, {"class": "bg-slate-800/50 backdrop-blur rounded-xl border border-purple-500/30 p-8 mb-20"}))
        return fragment(_root_children_3000)

