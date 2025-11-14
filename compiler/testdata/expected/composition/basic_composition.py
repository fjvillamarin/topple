from topple.psx import BaseView, Element, el, escape, fragment
class Button(BaseView):
    def __init__(self, text: str, variant: str="primary"):
        super().__init__()
        self.text = text
        self.variant = variant

    def _render(self) -> Element:
        return el("button", escape({self.text}), {"class": escape(f"btn btn-{self.variant}")})

class Card(BaseView):
    def __init__(self, title: str):
        super().__init__()
        self.title = title

    def _render(self) -> Element:
        return el("div", [el("div", el("h3", escape(self.title)), {"class": "card-header"}), el("div", el("p", "Card content goes here"), {"class": "card-body"})], {"class": "card"})

class App(BaseView):
    def __init__(self):
        super().__init__()

    def _render(self) -> Element:
        return el("div", [el("h1", "My Application"), Card(title="Welcome"), Button(text="Click me", variant="success"), Button(text="Cancel", variant="danger")], {"class": "app"})

