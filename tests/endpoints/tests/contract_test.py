import pytest
import requests
import assertpy
from utils.config import Devnet_URL, Qanet_URL, Testnet_URL, Mainnet_URL


@pytest.mark.parametrize("network", [Devnet_URL, Qanet_URL, Testnet_URL, Mainnet_URL])
def test_contract(network):
    response = requests.get(network+'contracts')
    response_text = response.json()

    assertpy.assert_that(response.status_code).is_equal_to(200)
    assertpy.assert_that(len(response_text)).is_equal_to(50)

    response = requests.get(network+'contracts?size=5')
    response_text = response.json()

    assertpy.assert_that(response.status_code).is_equal_to(200)
    assertpy.assert_that(len(response_text)).is_equal_to(5)

    response = requests.get(network+'contracts?contract_id=1')
    response_text = response.json()

    assertpy.assert_that(response.status_code).is_equal_to(200)
    assertpy.assert_that(response_text[0]).contains_key('contractId', 'twinId', 'state', 'created_at', 'type', 'details', 'billing')
    assertpy.assert_that(len(response_text)).is_equal_to(1)