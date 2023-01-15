import pytest
import requests
import assertpy
from utils.config import Devnet_URL, Qanet_URL, Testnet_URL, Mainnet_URL


@pytest.mark.parametrize("network", [Devnet_URL, Qanet_URL, Testnet_URL, Mainnet_URL])
def test_farm(network):
    response = requests.get(network+'farms')
    response_text = response.json()

    assertpy.assert_that(response.status_code).is_equal_to(200)
    assertpy.assert_that(len(response_text)).is_greater_than(1)

    response = requests.get(network+'farms?name=Freefarm')
    response_text = response.json()

    assertpy.assert_that(response.status_code).is_equal_to(200)
    assertpy.assert_that(response_text).is_not_empty()
    assertpy.assert_that(response_text[0]['name']).is_equal_to_ignoring_case('Freefarm')
    assertpy.assert_that(len(response_text[0])).is_greater_than_or_equal_to(1)

    response = requests.get(network+'farms?size=3')
    response_text = response.json()

    assertpy.assert_that(response.status_code).is_equal_to(200)
    assertpy.assert_that(len(response_text)).is_equal_to(3)