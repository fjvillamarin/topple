from topple.psx import BaseView, Element, el, escape, fragment, raw
class Icon(BaseView):
    def __init__(self, name: str):
        super().__init__()
        self.name = name

    def _render(self) -> Element:
        return el("i", "", {"class": escape(f"icon icon-{self.name}")})

class Button(BaseView):
    def __init__(self, text: str, icon: str=""):
        super().__init__()
        self.text = text
        self.icon = icon

    def _render(self) -> Element:
        _root_children_1000 = []
        _button_children_2000 = []
        if self.icon:
            _button_children_2000.append(Icon(name=self.icon))
        _button_children_2000.append(el("span", escape(self.text)))
        _root_children_1000.append(el("button", _button_children_2000))
        return fragment(_root_children_1000)

class Toolbar(BaseView):
    def __init__(self, title: str):
        super().__init__()
        self.title = title

    def _render(self) -> Element:
        _root_children_3000 = []
        _div_children_4000 = []
        _div_children_4000.append(el("h2", escape(self.title)))
        _div_children_5000 = []
        _div_children_5000.append(Button(text="Save", icon="save"))
        _div_children_5000.append(Button(text="Edit", icon="edit"))
        _div_children_5000.append(Button(text="Delete", icon="trash"))
        _div_children_4000.append(el("div", _div_children_5000, {"class": "actions"}))
        _root_children_3000.append(el("div", _div_children_4000, {"class": "toolbar"}))
        return fragment(_root_children_3000)

class Page(BaseView):
    def __init__(self):
        super().__init__()

    def _render(self) -> Element:
        _root_children_6000 = []
        _div_children_7000 = []
        _div_children_7000.append(Toolbar(title="Document Editor"))
        _div_children_7000.append(el("div", el("p", "Page content"), {"class": "content"}))
        _root_children_6000.append(el("div", _div_children_7000, {"class": "page"}))
        return fragment(_root_children_6000)

