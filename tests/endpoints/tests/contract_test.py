import requests
import assertpy
from utils.config import Devnet_URL, Qanet_URL, Testnet_URL, Mainnet_URL


def test_contract_devnet():
    response = requests.get(Devnet_URL+'contracts')
    response_text = response.json()

    assertpy.assert_that(response.status_code).is_equal_to(200)
    assertpy.assert_that(len(response_text)).is_equal_to(50)

    response = requests.get(Devnet_URL+'contracts?size=5')
    response_text = response.json()

    assertpy.assert_that(response.status_code).is_equal_to(200)
    assertpy.assert_that(len(response_text)).is_equal_to(5)

    response = requests.get(Devnet_URL+'contracts?contract_id=1')
    response_text = response.json()

    assertpy.assert_that(response.status_code).is_equal_to(200)
    assertpy.assert_that(response_text[0]).contains_key('contractId', 'twinId', 'state', 'created_at', 'type', 'details', 'billing')
    assertpy.assert_that(len(response_text)).is_equal_to(1)

def test_contract_qanet():
    response = requests.get(Qanet_URL+'contracts')
    response_text = response.json()

    assertpy.assert_that(response.status_code).is_equal_to(200)
    assertpy.assert_that(len(response_text)).is_equal_to(50)

    response = requests.get(Qanet_URL+'contracts?size=5')
    response_text = response.json()

    assertpy.assert_that(response.status_code).is_equal_to(200)
    assertpy.assert_that(len(response_text)).is_equal_to(5)

    response = requests.get(Qanet_URL+'contracts?contract_id=1')
    response_text = response.json()

    assertpy.assert_that(response.status_code).is_equal_to(200)
    assertpy.assert_that(response_text[0]).contains_key('contractId', 'twinId', 'state', 'created_at', 'type', 'details', 'billing')
    assertpy.assert_that(len(response_text)).is_equal_to(1)

def test_contract_testnet():
    response = requests.get(Testnet_URL+'contracts')
    response_text = response.json()

    assertpy.assert_that(response.status_code).is_equal_to(200)
    assertpy.assert_that(len(response_text)).is_equal_to(50)

    response = requests.get(Testnet_URL+'contracts?size=5')
    response_text = response.json()

    assertpy.assert_that(response.status_code).is_equal_to(200)
    assertpy.assert_that(len(response_text)).is_equal_to(5)

    response = requests.get(Testnet_URL+'contracts?contract_id=1')
    response_text = response.json()

    assertpy.assert_that(response.status_code).is_equal_to(200)
    assertpy.assert_that(response_text[0]).contains_key('contractId', 'twinId', 'state', 'created_at', 'type', 'details', 'billing')
    assertpy.assert_that(len(response_text)).is_equal_to(1)

def test_contract_mainnet():
    response = requests.get(Mainnet_URL+'contracts')
    response_text = response.json()

    assertpy.assert_that(response.status_code).is_equal_to(200)
    assertpy.assert_that(len(response_text)).is_equal_to(50)

    response = requests.get(Mainnet_URL+'contracts?size=5')
    response_text = response.json()

    assertpy.assert_that(response.status_code).is_equal_to(200)
    assertpy.assert_that(len(response_text)).is_equal_to(5)

    response = requests.get(Mainnet_URL+'contracts?contract_id=1')
    response_text = response.json()

    assertpy.assert_that(response.status_code).is_equal_to(200)
    assertpy.assert_that(response_text[0]).contains_key('contractId', 'twinId', 'state', 'created_at', 'type', 'details', 'billing')
    assertpy.assert_that(len(response_text)).is_equal_to(1)