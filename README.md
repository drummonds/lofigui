# lofigui

This is tooling for me as a go and python programmer to provide really simple front ends.  They serve the same area as:

It provides a way to build a very simple web app that can be bundled if required.
The use cases are:
- quick and simple
- more than a static website

The use cases are:
- providing a gui for a command line tool
- 1-10 users
- more for front ends for single physical object or a single process


I have used Bulma as a CSS framework to make it look prettier as I am terrible at design.

## Elements

- model view controller architecture
- templating 
- style sheets
- buffering

Your project is essentially a web site.  To make design simple you completely refresh pages so no code for partial refreshes.  To make things dynamic it has to be asynchonous so for python using fastapi as a server and Uvicorn to provide the https server.

Like a normal terminal program you essentially just print things to a screen but now have the ability to print enriched objects.

### model view controller architecture
All I really want to do is to write the model.  The controller and view (in the browser and templating system) are a necessary evil.  The controller includes the routing and webserver.  The view is the html templating and the browser.

### Buffer
In order to be able to decouple the display from the output and to be able to refesh you need to be able to buffer the output.  It is more efficient to buffer the output in the browser but more complicated.  Moving the buffer to the server simplifies the software but requires you to refresh the whole page.


## Alternative approaches

- [pywebio](https://www.pyweb.io/)
- [streamlit](https://streamlit.io/)
- [textual](https://pypi.org/project/textual/)

The difference is that this approach should be very simple and easily understandable.
For the moment no Javascript is used.


## Roadmap

- A go version, will be event simpler
- A go wasm version for deploying serverless (no physical object)