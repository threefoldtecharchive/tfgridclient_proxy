import requests
import assertpy
from utils.config import Devnet_URL, Qanet_URL, Testnet_URL, Mainnet_URL


def test_twin_devnet():
    response = requests.get(Devnet_URL+'twins')
    response_text = response.json()
    
    assertpy.assert_that(response.status_code).is_equal_to(200)
    assertpy.assert_that(len(response_text)).is_greater_than(1)
    
    response = requests.get(Devnet_URL+'twins?twin_id=1')
    response_text = response.json()
    
    assertpy.assert_that(response.status_code).is_equal_to(200)
    assertpy.assert_that(response_text).does_not_contain('None', 'null')
    assertpy.assert_that(len(response_text[0])).is_greater_than(1)

def test_twin_qanet():
    response = requests.get(Qanet_URL+'twins')
    response_text = response.json()
    
    assertpy.assert_that(response.status_code).is_equal_to(200)
    assertpy.assert_that(len(response_text)).is_greater_than(1)
    
    response = requests.get(Qanet_URL+'twins?twin_id=1')
    response_text = response.json()
    print(response_text)
    
    assertpy.assert_that(response.status_code).is_equal_to(200)
    assertpy.assert_that(response_text).does_not_contain('None')
    assertpy.assert_that(len(response_text[0])).is_greater_than(1)

def test_twin_testnet():
    response = requests.get(Testnet_URL+'twins')
    response_text = response.json()
    
    assertpy.assert_that(response.status_code).is_equal_to(200)
    assertpy.assert_that(len(response_text)).is_greater_than(1)
    
    response = requests.get(Testnet_URL+'twins?twin_id=1')
    response_text = response.json()
    print(response_text)
    
    assertpy.assert_that(response.status_code).is_equal_to(200)
    assertpy.assert_that(response_text).does_not_contain('None', 'null')
    assertpy.assert_that(len(response_text[0])).is_greater_than(1)

def test_twin_mainnet():
    response = requests.get(Mainnet_URL+'twins')
    response_text = response.json()
    
    assertpy.assert_that(response.status_code).is_equal_to(200)
    assertpy.assert_that(len(response_text)).is_greater_than(1)
    
    response = requests.get(Mainnet_URL+'twins?twin_id=1')
    response_text = response.json()
    
    assertpy.assert_that(response.status_code).is_equal_to(200)
    assertpy.assert_that(response_text).does_not_contain('None', 'null')
    assertpy.assert_that(len(response_text[0])).is_greater_than(1)