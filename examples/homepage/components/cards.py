from topple.psx import BaseView, Element, el, escape, fragment, raw
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
        _div_children_4000.append(el("pre", el("code", escape(raw(self.code))), {"class": "text-slate-300 text-xs font-mono whitespace-pre overflow-x-auto"}))
        _root_children_3000.append(el("div", _div_children_4000, {"class": "bg-slate-900 rounded-lg p-4"}))
        return fragment(_root_children_3000)

