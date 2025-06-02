# Biscuit: Pythonic Templating for Modern Web UIs

Biscuit is an innovative templating language that seamlessly blends the power and familiarity of Python with standard HTML markup. It's designed for developers who want to leverage Python's full programming capabilities to create dynamic and interactive user interfaces, offering a development experience conceptually similar to JSX/TSX in the JavaScript world.

With Biscuit, you can write expressive, reusable UI components (called "views") using Python logic directly within your HTML structures. This allows for clean separation of concerns while keeping related display logic and markup colocated.

## Key Features & Benefits

*   **Python-Powered Templating**: Use all of Python's features directly in your templates – variables, loops, conditionals, functions, classes, and more. No need to learn a separate, restricted templating DSL.
*   **Reusable UI Components (Views)**: Build your UI as a collection of modular, composable views. Pass data as parameters and manage state with Python's natural constructs.
*   **Seamless FastAPI Integration**: Biscuit is designed to work effortlessly with FastAPI. Views can be returned directly from your route handlers, simplifying the development of full-stack Python web applications.
*   **First-Class HTMX Support**: Create rich, interactive user experiences with minimal JavaScript. Biscuit provides direct support for HTMX attributes, making it easy to build modern, dynamic UIs.
*   **Vue-Inspired Slot System**: Enjoy flexible and powerful layout capabilities with a slot system inspired by Vue.js, allowing for sophisticated content injection and component composition.

## Quick Example

Here's a taste of what a Biscuit view looks like:

```python
# hello_view.bsct
view Greeting(name: str = "World"):
    <div class="greeting">
        <h1>Hello, {name}!</h1>
        if len(name) > 10:
            <p>That's a long name!</p>
        else:
            <p>Nice to meet you.</p>
    </div>
```

## Learn More

For a comprehensive guide to the Biscuit language, its syntax, features, and advanced usage, please see the [Biscuit Language Grammar](grammar_biscuit.md) documentation.

Dive into the `examples/` directory in this repository to see Biscuit in action with various use cases, including FastAPI integration and HTMX examples.
