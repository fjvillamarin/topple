from topple.psx import BaseView, Element, el, escape, fragment
from .card import Card
class Page(BaseView):
    def __init__(self):
        super().__init__()

    def _render(self) -> Element:
        _root_children_1000 = []
        _div_children_2000 = []
        _div_children_2000.append(el("h1", "My Page"))
        _div_children_2000.append(Card(title="First Card"))
        _div_children_2000.append(Card(title="Second Card"))
        _root_children_1000.append(el("div", _div_children_2000))
        return fragment(_root_children_1000)

