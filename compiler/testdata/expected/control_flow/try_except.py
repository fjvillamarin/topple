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
        _root_children_2000 = []
        _div_children_3000 = []
        try:
            result = risky_operation(self.input_value)
            _div_children_3000.append(el("div", [el("h2", "Success!"), el("p", f"Result: {escape(result)}")], {"class": "success"}))
        except ValueError as e:
            _div_children_3000.append(el("div", [el("h2", "Error"), el("p", f"ValueError: {escape(str(e))}")], {"class": "error"}))
        except Exception as e:
            _div_children_3000.append(el("div", [el("h2", "Unexpected Error"), el("p", f"Error: {escape(str(e))}")], {"class": "error"}))
        finally:
            _div_children_3000.append(el("p", "Operation completed", {"class": "footer"}))
        _root_children_2000.append(el("div", _div_children_3000))
        return fragment(_root_children_2000)

