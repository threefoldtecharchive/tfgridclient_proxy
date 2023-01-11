import requests
import assertpy
from utils.config import Devnet_URL, Qanet_URL, Testnet_URL, Mainnet_URL


def test_gateway_devnet():
    response = requests.get(Devnet_URL+'gateways')
    response_text = response.json()

    assertpy.assert_that(response.status_code).is_equal_to(200)
    assertpy.assert_that(len(response_text)).is_greater_than(1)

    response = requests.get(Devnet_URL+'gateways?farm_ids=1')
    response_text = response.json()

    assertpy.assert_that(response.status_code).is_equal_to(200)
    assertpy.assert_that(response_text).is_not_empty()
    assertpy.assert_that(len(response_text[0])).is_greater_than(1)

    response = requests.get(Devnet_URL+'gateways')
    response_text = response.json()
    response = requests.get(Devnet_URL+'gateways/'+str(response_text[0]['nodeId']))
    response_text = response.json()

    assertpy.assert_that(response.status_code).is_equal_to(200)
    assertpy.assert_that(response_text).contains_key("capacity", "certificationType", "city", "country", "created", "dedicated", "farmId", "farmingPolicyId",
                                                     "gridVersion", "id", "location", "nodeId", "publicConfig", "rentContractId", "rentedByTwinId", "serialNumber", "status", "twinId", "updatedAt", "uptime")
    assertpy.assert_that(len(response_text)).is_greater_than(1)


def test_gateway_qanet():
    response = requests.get(Qanet_URL+'gateways')
    response_text = response.json()

    assertpy.assert_that(response.status_code).is_equal_to(200)
    assertpy.assert_that(len(response_text)).is_greater_than(1)

    response = requests.get(Qanet_URL+'gateways?farm_ids=1')
    response_text = response.json()

    assertpy.assert_that(response.status_code).is_equal_to(200)
    assertpy.assert_that(response_text).is_not_empty()
    assertpy.assert_that(len(response_text[0])).is_greater_than(1)

    response = requests.get(Qanet_URL+'gateways')
    response_text = response.json()
    response = requests.get(Qanet_URL+'gateways/'+str(response_text[0]['nodeId']))
    response_text = response.json()

    assertpy.assert_that(response.status_code).is_equal_to(200)
    assertpy.assert_that(response_text).contains_key("capacity", "certificationType", "city", "country", "created", "dedicated", "farmId", "farmingPolicyId",
                                                     "gridVersion", "id", "location", "nodeId", "publicConfig", "rentContractId", "rentedByTwinId", "serialNumber", "status", "twinId", "updatedAt", "uptime")
    assertpy.assert_that(len(response_text)).is_greater_than(1)


def test_gateway_testnet():
    response = requests.get(Testnet_URL+'gateways')
    response_text = response.json()

    assertpy.assert_that(response.status_code).is_equal_to(200)
    assertpy.assert_that(len(response_text)).is_greater_than(1)

    response = requests.get(Testnet_URL+'gateways?farm_ids=1')
    response_text = response.json()

    assertpy.assert_that(response.status_code).is_equal_to(200)
    assertpy.assert_that(response_text).is_not_empty()
    assertpy.assert_that(len(response_text[0])).is_greater_than(1)

    response = requests.get(Testnet_URL+'gateways')
    response_text = response.json()
    response = requests.get(Testnet_URL+'gateways/'+str(response_text[0]['nodeId']))
    response_text = response.json()

    assertpy.assert_that(response.status_code).is_equal_to(200)
    assertpy.assert_that(response_text).contains_key("capacity", "certificationType", "city", "country", "created", "dedicated", "farmId", "farmingPolicyId",
                                                     "gridVersion", "id", "location", "nodeId", "publicConfig", "rentContractId", "rentedByTwinId", "serialNumber", "status", "twinId", "updatedAt", "uptime")
    assertpy.assert_that(len(response_text)).is_greater_than(1)


def test_gateway_mainnet():
    response = requests.get(Mainnet_URL+'gateways')
    response_text = response.json()

    assertpy.assert_that(response.status_code).is_equal_to(200)
    assertpy.assert_that(len(response_text)).is_greater_than(1)

    response = requests.get(Mainnet_URL+'gateways?farm_ids=1')
    response_text = response.json()

    assertpy.assert_that(response.status_code).is_equal_to(200)
    assertpy.assert_that(response_text).is_not_empty()
    assertpy.assert_that(len(response_text[0])).is_greater_than(1)

    response = requests.get(Mainnet_URL+'gateways')
    response_text = response.json()
    response = requests.get(Mainnet_URL+'gateways/'+str(response_text[0]['nodeId']))
    response_text = response.json()

    assertpy.assert_that(response.status_code).is_equal_to(200)
    assertpy.assert_that(response_text).contains_key("capacity", "certificationType", "city", "country", "created", "dedicated", "farmId", "farmingPolicyId",
                                                     "gridVersion", "id", "location", "nodeId", "publicConfig", "rentContractId", "rentedByTwinId", "serialNumber", "status", "twinId", "updatedAt", "uptime")
    assertpy.assert_that(len(response_text)).is_greater_than(1)
