from topple.psx import BaseView, Element, el, escape, fragment, raw
class Button(BaseView):
    def __init__(self, text: str, variant: str="primary"):
        super().__init__()
        self.text = text
        self.variant = variant

    def _render(self) -> Element:
        return el("button", escape(self.text), {"class": escape(f"btn btn-{self.variant}")})

class Card(BaseView):
    def __init__(self, title: str):
        super().__init__()
        self.title = title

    def _render(self) -> Element:
        _root_children_1000 = []
        _div_children_2000 = []
        _div_children_2000.append(el("div", el("h3", escape(self.title)), {"class": "card-header"}))
        _div_children_2000.append(el("div", el("p", "Card content goes here"), {"class": "card-body"}))
        _root_children_1000.append(el("div", _div_children_2000, {"class": "card"}))
        return fragment(_root_children_1000)

class App(BaseView):
    def __init__(self):
        super().__init__()

    def _render(self) -> Element:
        _root_children_3000 = []
        _div_children_4000 = []
        _div_children_4000.append(el("h1", "My Application"))
        _div_children_4000.append(Card(title="Welcome"))
        _div_children_4000.append(Button(text="Click me", variant="success"))
        _div_children_4000.append(Button(text="Cancel", variant="danger"))
        _root_children_3000.append(el("div", _div_children_4000, {"class": "app"}))
        return fragment(_root_children_3000)

