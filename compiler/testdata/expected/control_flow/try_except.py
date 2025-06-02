from runtime import BaseView, el, escape, Element, FragmentElement, fragment, render_child
def risky_operation(value):
    if value < 0:
        raise ValueError("Negative value not allowed")
    return value * 2

class ErrorHandlingView(BaseView):
    def __init__(self, input_value: int):
        super().__init__()
        self.input_value = input_value

    def _render(self) -> Element:
        _div_children_1000 = []
        try:
            result = risky_operation(input_value)
            _div_children_1000.append(el("div", [el("h2", "Success!"), el("p", f"Result:{escape(result)}")], {"class": "success"}))
        except ValueError as e:
            _div_children_1000.append(el("div", [el("h2", "Error"), el("p", f"ValueError:{escape(str(e))}")], {"class": "error"}))
        except Exception as e:
            _div_children_1000.append(el("div", [el("h2", "Unexpected Error"), el("p", f"Error:{escape(str(e))}")], {"class": "error"}))
        finally:
            _div_children_1000.append(el("p", "Operation completed", {"class": "footer"}))
        return el("div", _div_children_1000)

