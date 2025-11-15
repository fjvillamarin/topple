from topple.psx import BaseView, Element, el, escape, fragment
from ..components.cards import ExampleCard
from ..utils.data import EXAMPLES
class ExamplesSection(BaseView):
    def __init__(self):
        super().__init__()

    def _render(self) -> Element:
        _root_children_1000 = []
        _div_children_2000 = []
        _div_children_2000.append(el("h2", "Works with Compound Statements", {"class": "text-3xl font-bold text-purple-300 mb-6 text-center"}))
        _div_children_3000 = []
        for example in EXAMPLES:
            _div_children_3000.append(ExampleCard(label=example["label"], code=example["code"]))
        _div_children_2000.append(el("div", _div_children_3000, {"class": "grid md:grid-cols-3 gap-6"}))
        _root_children_1000.append(el("div", _div_children_2000, {"class": "bg-slate-800/50 backdrop-blur rounded-xl border border-purple-500/30 p-8"}))
        return fragment(_root_children_1000)

