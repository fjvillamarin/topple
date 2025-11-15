from topple.psx import BaseView, Element, el, escape, fragment
class MatchView(BaseView):
    def __init__(self, status: str, data: dict):
        super().__init__()
        self.status = status
        self.data = data

    def _render(self) -> Element:
        _root_children_1000 = []
        _div_children_2000 = []
        match self.status:
            case "loading":
                _div_children_2000.append(el("div", "Loading...", {"class": "spinner"}))
            case "success":
                _div_children_3000 = []
                _div_children_3000.append(el("h2", "Success!"))
                _div_children_3000.append(el("p", f"Data loaded: {escape(len(self.data))} items"))
                _div_children_2000.append(el("div", _div_children_3000, {"class": "success"}))
            case "error":
                _div_children_4000 = []
                _div_children_4000.append(el("h2", "Error occurred"))
                _div_children_4000.append(el("p", "Failed to load data"))
                _div_children_2000.append(el("div", _div_children_4000, {"class": "error"}))
            case _:
                _div_children_2000.append(el("div", el("p", f"Unknown status: {escape(self.status)}"), {"class": "unknown"}))
        _root_children_1000.append(el("div", _div_children_2000))
        return fragment(_root_children_1000)

