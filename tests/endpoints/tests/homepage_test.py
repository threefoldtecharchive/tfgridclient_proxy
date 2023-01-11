import requests
import assertpy
from utils.config import Devnet_URL, Qanet_URL, Testnet_URL, Mainnet_URL


def test_homepage_devnet():
    response = requests.get(Devnet_URL)
    response_text = response.json()

    assertpy.assert_that(response.status_code).is_equal_to(200)
    assertpy.assert_that(response_text).contains("Welcome to threefold grid proxy server")

def test_homepage_qanet():
    response = requests.get(Qanet_URL)
    response_text = response.json()

    assertpy.assert_that(response.status_code).is_equal_to(200)
    assertpy.assert_that(response_text).contains("Welcome to threefold grid proxy server")

def test_homepage_testnet():
    response = requests.get(Testnet_URL)
    response_text = response.json()

    assertpy.assert_that(response.status_code).is_equal_to(200)
    assertpy.assert_that(response_text).contains("welcome to grid proxy server")

def test_homepage_mainnet():
    response = requests.get(Mainnet_URL)
    response_text = response.json()

    assertpy.assert_that(response.status_code).is_equal_to(200)
    assertpy.assert_that(response_text).contains("welcome to grid proxy server")