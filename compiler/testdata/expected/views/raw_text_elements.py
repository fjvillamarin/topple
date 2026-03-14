from topple.psx import BaseView, Element, el, escape, fragment, raw
class ScriptString(BaseView):
    def __init__(self):
        super().__init__()

    def _render(self) -> Element:
        return el("script", "\"alert('hello')\"")

class ScriptInterpolation(BaseView):
    def __init__(self, config: str):
        super().__init__()
        self.config = config

    def _render(self) -> Element:
        return el("script", raw(self.config))

class StyleString(BaseView):
    def __init__(self):
        super().__init__()

    def _render(self) -> Element:
        return el("style", "\"div > p, span + a\"")

class StyleInterpolation(BaseView):
    def __init__(self, color: str):
        super().__init__()
        self.color = color

    def _render(self) -> Element:
        return el("style", raw(self.color))

class TextareaView(BaseView):
    def __init__(self, value: str):
        super().__init__()
        self.value = value

    def _render(self) -> Element:
        return el("textarea", raw(self.value))

class MixedView(BaseView):
    def __init__(self, name: str):
        super().__init__()
        self.name = name

    def _render(self) -> Element:
        _root_children_1000 = []
        _root_children_1000.append(el("div", f"Hello, {escape(self.name)}!"))
        _root_children_1000.append(el("script", "\"console.log('hello')\""))
        return fragment(_root_children_1000)

