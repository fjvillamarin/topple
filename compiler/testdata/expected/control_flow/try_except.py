from topple.psx import BaseView, Element, el, escape, fragment, raw
def risky_operation(value):
    if value < 0:
        raise ValueError("Negative value not allowed")
    return value * 2

class ErrorHandlingView(BaseView):
    def __init__(self, input_value: int):
        super().__init__()
        self.input_value = input_value

    def _render(self) -> Element:
        _root_children_1000 = []
        _div_children_2000 = []
        try:
            result = risky_operation(self.input_value)
            _div_children_3000 = []
            _div_children_3000.append(el("h2", "Success!"))
            _div_children_3000.append(el("p", f"Result: {escape(result)}"))
            _div_children_2000.append(el("div", _div_children_3000, {"class": "success"}))
        except ValueError as e:
            _div_children_4000 = []
            _div_children_4000.append(el("h2", "Error"))
            _div_children_4000.append(el("p", f"ValueError: {escape(str(e))}"))
            _div_children_2000.append(el("div", _div_children_4000, {"class": "error"}))
        except Exception as e:
            _div_children_5000 = []
            _div_children_5000.append(el("h2", "Unexpected Error"))
            _div_children_5000.append(el("p", f"Error: {escape(str(e))}"))
            _div_children_2000.append(el("div", _div_children_5000, {"class": "error"}))
        finally:
            _div_children_2000.append(el("p", "Operation completed", {"class": "footer"}))
        _root_children_1000.append(el("div", _div_children_2000))
        return fragment(_root_children_1000)

