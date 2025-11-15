from topple.psx import BaseView, Element, el, escape, fragment
class ValidationErrors(BaseView):
    def __init__(self, errors: dict):
        super().__init__()
        self.errors = errors

    def _render(self) -> Element:
        _root_children_1000 = []
        _div_children_2000 = []
        _div_children_2000.append(el("h3", "Please correct the following errors:"))
        _ul_children_3000 = []
        for (field, error) in self.errors.items():
            _ul_children_3000.append(el("li", f"{escape(field)}: {escape(error)}"))
        _div_children_2000.append(el("ul", _ul_children_3000))
        _root_children_1000.append(el("div", _div_children_2000, {"class": "error-banner"}))
        return fragment(_root_children_1000)

class ContactForm(BaseView):
    def __init__(self):
        super().__init__()

    def _render(self) -> Element:
        _root_children_4000 = []
        _div_children_5000 = []
        _div_children_5000.append(el("h1", "Contact Us"))
        _form_children_6000 = []
        _div_children_7000 = []
        _div_children_7000.append(el("label", "Name *", {"for": "name"}))
        _div_children_7000.append(el("input", "", {"type": "text", "id": "name", "name": "name", "required": True, "hx-post": "/api/validate/name", "hx-trigger": "blur", "hx-target": "next .validation-message"}))
        _div_children_7000.append(el("div", "", {"class": "validation-message"}))
        _form_children_6000.append(el("div", _div_children_7000, {"class": "form-group"}))
        _div_children_8000 = []
        _div_children_8000.append(el("label", "Email *", {"for": "email"}))
        _div_children_8000.append(el("input", "", {"type": "email", "id": "email", "name": "email", "required": True, "hx-post": "/api/validate/email", "hx-trigger": "blur", "hx-target": "next .validation-message"}))
        _div_children_8000.append(el("div", "", {"class": "validation-message"}))
        _form_children_6000.append(el("div", _div_children_8000, {"class": "form-group"}))
        _div_children_9000 = []
        _div_children_9000.append(el("label", "Message *", {"for": "message"}))
        _div_children_9000.append(el("textarea", "", {"id": "message", "name": "message", "required": True, "hx-post": "/api/validate/message", "hx-trigger": "blur", "hx-target": "next .validation-message"}))
        _div_children_9000.append(el("div", "", {"class": "validation-message"}))
        _form_children_6000.append(el("div", _div_children_9000, {"class": "form-group"}))
        _form_children_6000.append(el("button", escape("Send Message"), {"type": "submit", "hx-indicator": ".submit-spinner"}))
        _form_children_6000.append(el("span", "Sending...", {"class": "submit-spinner", "style": "display:none;"}))
        _div_children_5000.append(el("form", _form_children_6000, {"hx-post": "/api/contact", "hx-swap": "outerHTML", "hx-target": "this"}))
        _root_children_4000.append(el("div", _div_children_5000))
        return fragment(_root_children_4000)

