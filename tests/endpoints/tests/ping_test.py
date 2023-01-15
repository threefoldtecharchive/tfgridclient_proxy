import pytest
import requests
import assertpy
from utils.config import Devnet_URL, Qanet_URL, Testnet_URL, Mainnet_URL


@pytest.mark.parametrize("network", [Devnet_URL, Qanet_URL, Testnet_URL, Mainnet_URL])
def test_ping(network):
    response = requests.get(network+'ping')
    response_text = response.json()

    assertpy.assert_that(response.status_code).is_equal_to(200)
    assertpy.assert_that(response_text).contains_entry({"ping": "pong"})