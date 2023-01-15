import pytest
import requests
import assertpy
from utils.config import Devnet_URL, Qanet_URL, Testnet_URL, Mainnet_URL


@pytest.mark.parametrize("network", [Devnet_URL, Qanet_URL, Testnet_URL, Mainnet_URL])
def test_twin(network):
    response = requests.get(network+'twins')
    response_text = response.json()
    
    assertpy.assert_that(response.status_code).is_equal_to(200)
    assertpy.assert_that(len(response_text)).is_greater_than(1)
    
    response = requests.get(network+'twins?twin_id=1')
    response_text = response.json()
    
    assertpy.assert_that(response.status_code).is_equal_to(200)
    assertpy.assert_that(response_text).does_not_contain('None', 'null')
    assertpy.assert_that(len(response_text[0])).is_greater_than(1)