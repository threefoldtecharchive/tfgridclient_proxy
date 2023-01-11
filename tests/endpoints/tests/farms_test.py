import requests
import assertpy
from utils.config import Devnet_URL, Qanet_URL, Testnet_URL, Mainnet_URL


def test_farm_devnet():
    response = requests.get(Devnet_URL+'farms')
    response_text = response.json()

    assertpy.assert_that(response.status_code).is_equal_to(200)
    assertpy.assert_that(len(response_text)).is_greater_than(1)

    response = requests.get(Devnet_URL+'farms?name=Freefarm')
    response_text = response.json()

    assertpy.assert_that(response.status_code).is_equal_to(200)
    assertpy.assert_that(response_text).is_not_empty()
    assertpy.assert_that(response_text[0]['name']).is_equal_to('Freefarm')
    assertpy.assert_that(len(response_text[0])).is_greater_than_or_equal_to(1)

    response = requests.get(Devnet_URL+'farms?size=3')
    response_text = response.json()

    assertpy.assert_that(response.status_code).is_equal_to(200)
    assertpy.assert_that(len(response_text)).is_equal_to(3)

def test_farm_qanet():
    response = requests.get(Qanet_URL+'farms')
    response_text = response.json()

    assertpy.assert_that(response.status_code).is_equal_to(200)
    assertpy.assert_that(len(response_text)).is_greater_than(1)

    response = requests.get(Qanet_URL+'farms?name=Freefarm')
    response_text = response.json()

    assertpy.assert_that(response.status_code).is_equal_to(200)
    assertpy.assert_that(response_text).is_not_empty()
    assertpy.assert_that(response_text[0]['name']).is_equal_to('Freefarm')
    assertpy.assert_that(len(response_text[0])).is_greater_than_or_equal_to(1)

    response = requests.get(Qanet_URL+'farms?size=3')
    response_text = response.json()

    assertpy.assert_that(response.status_code).is_equal_to(200)
    assertpy.assert_that(len(response_text)).is_equal_to(3)

def test_farm_testnet():
    response = requests.get(Testnet_URL+'farms')
    response_text = response.json()

    assertpy.assert_that(response.status_code).is_equal_to(200)
    assertpy.assert_that(len(response_text)).is_greater_than(1)

    response = requests.get(Testnet_URL+'farms?name=Freefarm')
    response_text = response.json()

    assertpy.assert_that(response.status_code).is_equal_to(200)
    assertpy.assert_that(response_text).is_not_empty()
    assertpy.assert_that(response_text[0]['name']).is_equal_to('FreeFarm')
    assertpy.assert_that(len(response_text[0])).is_greater_than_or_equal_to(1)

    response = requests.get(Testnet_URL+'farms?size=3')
    response_text = response.json()

    assertpy.assert_that(response.status_code).is_equal_to(200)
    assertpy.assert_that(len(response_text)).is_equal_to(3)

def test_farm_mainnet():
    response = requests.get(Mainnet_URL+'farms')
    response_text = response.json()

    assertpy.assert_that(response.status_code).is_equal_to(200)
    assertpy.assert_that(len(response_text)).is_greater_than(1)

    response = requests.get(Mainnet_URL+'farms?name=Freefarm')
    response_text = response.json()

    assertpy.assert_that(response.status_code).is_equal_to(200)
    assertpy.assert_that(response_text).is_not_empty()
    assertpy.assert_that(response_text[0]['name']).is_equal_to('Freefarm')
    assertpy.assert_that(len(response_text[0])).is_greater_than_or_equal_to(1)

    response = requests.get(Mainnet_URL+'farms?size=3')
    response_text = response.json()

    assertpy.assert_that(response.status_code).is_equal_to(200)
    assertpy.assert_that(len(response_text)).is_equal_to(3)