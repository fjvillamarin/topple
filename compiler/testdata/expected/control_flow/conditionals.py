from runtime import BaseView, el, escape, Element, FragmentElement, fragment, render_child
class ConditionalView(BaseView):
    def __init__(self, user_type: str, is_admin: bool = False):
        super().__init__()
        self.user_type = user_type
        self.is_admin = is_admin

    def _render(self) -> Element:
        _div_children_1000 = []
        if self.user_type == "guest":
            _div_children_1000.append(el("p", "Welcome, guest!"))
        else:
            if self.user_type == "user":
                _div_children_1000.append(el("p", "Hello, registered user!"))
            else:
                _div_children_1000.append(el("p", f"Welcome,{escape(self.user_type)}!"))
        if self.is_admin:
            _div_children_1000.append(el("div", [el("h3", "Admin Controls"), el("button", "Admin Actions")], {"class": "admin-panel"}))
        else:
            _div_children_1000.append(el("p", "Regular user view"))
        return el("div", _div_children_1000)

