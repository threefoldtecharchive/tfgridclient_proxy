import os
import pytest
import schemathesis
from schemathesis.checks import not_a_server_error, status_code_conformance, content_type_conformance, response_schema_conformance, response_headers_conformance

network = os.environ['NETWORK']
if network == 'main':
    url = 'https://gridproxy.grid.tf'
else:
    url = 'https://gridproxy.' + network + '.grid.tf'
    
schema = schemathesis.from_path("docs/swagger.json", base_url = url)


@pytest.mark.parametrize("check", [not_a_server_error, status_code_conformance, content_type_conformance, response_schema_conformance, response_headers_conformance])
@schema.parametrize()
def test_api(case, check):
    response = case.call()
    case.validate_response(response, checks=(check,))
