# Endpoints Test

Install the recommended version of the pip package listed below for a stable run, or you can just install python 3 and use `pip install -r requirements.txt` in the E2E directory.

| Prerequisites                                                    | version  |
| ---------------------------------------------------------------- | -------- |
| [Python](https://www.python.org/downloads/)                      | `3.10.4` |
| [pytest](https://pypi.org/project/pytest/)                       | `7.1.2`  |
| [requests](https://pypi.org/project/requests/)                   | `2.28.1` |
| [assertpy](https://pypi.org/project/assertpy/)                   | `1.1`    |

# Running tests

- Change direcotry to endpoints through the command line using `cd tests/endpoints`
- You can run selenium with pytest through the command line using `python3 -m pytest -v`

### More options to run tests

- You can also run single test file through the command line using `python3 -m pytest tests/test_file.py`
- You can also run specific test case through the command line using `python3 -m pytest tests/test_file.py::test_func`
- You can also run collection of test cases through the command line using `python3 -m pytest -v -k 'test_func or test_func'`
- You can also run all the tests and get an HTML report using [pytest-html](https://pypi.org/project/pytest-html/) package through the command line using `python3 -m pytest -v --html=report.html`
