import pytest
import schemathesis
from schemathesis.checks import not_a_server_error, status_code_conformance, content_type_conformance, response_schema_conformance, response_headers_conformance

schema = schemathesis.from_path("docs/swagger.json", base_url='https://gridproxy.dev.grid.tf')

@pytest.mark.parametrize("check", [not_a_server_error, status_code_conformance, content_type_conformance, response_schema_conformance, response_headers_conformance])
@schema.parametrize()
def test_api(case, check):
    response = case.call()
    case.validate_response(response, checks=(check,))
