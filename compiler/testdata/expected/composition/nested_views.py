from topple.psx import BaseView, Element, el, escape, fragment
class Icon(BaseView):
    def __init__(self, name: str):
        super().__init__()
        self.name = name

    def _render(self) -> Element:
        return el("i", "", {"class": "icon icon-{name}"})

class Button(BaseView):
    def __init__(self, text: str, icon: str=""):
        super().__init__()
        self.text = text
        self.icon = icon

    def _render(self) -> Element:
        return el("button", ["", el("span", escape(self.text))])

class Toolbar(BaseView):
    def __init__(self, title: str):
        super().__init__()
        self.title = title

    def _render(self) -> Element:
        return el("div", [el("h2", escape(self.title)), el("div", [Button(text="Save", icon="save"), Button(text="Edit", icon="edit"), Button(text="Delete", icon="trash")], {"class": "actions"})], {"class": "toolbar"})

class Page(BaseView):
    def __init__(self):
        super().__init__()

    def _render(self) -> Element:
        return el("div", [Toolbar(title="Document Editor"), el("div", el("p", "Page content"), {"class": "content"})], {"class": "page"})

