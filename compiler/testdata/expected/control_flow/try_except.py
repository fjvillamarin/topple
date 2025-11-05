from topple.psx import BaseView, Element, el, escape, fragment
def risky_operation(value):
    if value < 0:
        raise ValueError("Negative value not allowed")
    return value * 2

class ErrorHandlingView(BaseView):
    def __init__(self, input_value: int):
        super().__init__()
        self.input_value = input_value

    def _render(self) -> Element:
        return el("div", "")

