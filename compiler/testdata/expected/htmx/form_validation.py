from psx.runtime import BaseView, Element, el, escape, fragment
class ValidationErrors(BaseView):
    def __init__(self, errors: dict):
        super().__init__()
        self.errors = errors

    def _render(self) -> Element:
        return el("div", [el("h3", "Please correct the following errors:"), el("ul", "")], {"class": "error-banner"})

class ContactForm(BaseView):
    def __init__(self):
        super().__init__()

    def _render(self) -> Element:
        return el("div", [el("h1", "Contact Us"), el("form", [el("div", [el("label", "Name *", {"for": "name"}), el("input", "", {"type": "text", "id": "name", "name": "name", "required": True, "hx-post": "/api/validate/name", "hx-trigger": "blur", "hx-target": "next .validation-message"}), el("div", "", {"class": "validation-message"})], {"class": "form-group"}), el("div", [el("label", "Email *", {"for": "email"}), el("input", "", {"type": "email", "id": "email", "name": "email", "required": True, "hx-post": "/api/validate/email", "hx-trigger": "blur", "hx-target": "next .validation-message"}), el("div", "", {"class": "validation-message"})], {"class": "form-group"}), el("div", [el("label", "Message *", {"for": "message"}), el("textarea", "", {"id": "message", "name": "message", "required": True, "hx-post": "/api/validate/message", "hx-trigger": "blur", "hx-target": "next .validation-message"}), el("div", "", {"class": "validation-message"})], {"class": "form-group"}), el("button", ["", ""], {"type": "submit", "hx-indicator": ".submit-spinner"}), el("span", "Sending...", {"class": "submit-spinner", "style": "display:none;"})], {"hx-post": "/api/contact", "hx-swap": "outerHTML", "hx-target": "this"})])

