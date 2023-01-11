import requests
import assertpy
from utils.config import Devnet_URL, Qanet_URL, Testnet_URL, Mainnet_URL


def test_ping_devnet():
    response = requests.get(Devnet_URL+'ping')
    response_text = response.json()

    assertpy.assert_that(response.status_code).is_equal_to(200)
    assertpy.assert_that(response_text).contains_entry({"ping": "pong"})

def test_ping_qanet():
    response = requests.get(Qanet_URL+'ping')
    response_text = response.json()

    assertpy.assert_that(response.status_code).is_equal_to(200)
    assertpy.assert_that(response_text).contains_entry({"ping": "pong"})

def test_ping_testnet():
    response = requests.get(Testnet_URL+'ping')
    response_text = response.json()

    assertpy.assert_that(response.status_code).is_equal_to(200)
    assertpy.assert_that(response_text).contains_entry({"ping": "pong"})

def test_ping_mainnet():
    response = requests.get(Mainnet_URL+'ping')
    response_text = response.json()

    assertpy.assert_that(response.status_code).is_equal_to(200)
    assertpy.assert_that(response_text).contains_entry({"ping": "pong"})