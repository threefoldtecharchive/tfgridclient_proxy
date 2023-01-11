import requests
import assertpy
from utils.config import Devnet_URL, Qanet_URL, Testnet_URL, Mainnet_URL


def test_stats_devnet():
    response = requests.get(Devnet_URL+'stats?status=up')
    response_text = response.json()

    assertpy.assert_that(response.status_code).is_equal_to(200)
    assertpy.assert_that(response_text).contains_key("nodes", 'farms', 'countries', 'totalCru', 'totalSru',
                                                     'totalMru', 'totalHru', 'publicIps', 'accessNodes', 'gateways', 'twins', 'contracts', 'nodesDistribution')
    assertpy.assert_that(len(response_text['nodesDistribution'])).is_greater_than_or_equal_to(1)

    response = requests.get(Devnet_URL+'stats?status=down')
    response_text = response.json()

    assertpy.assert_that(response.status_code).is_equal_to(200)
    assertpy.assert_that(response_text).contains_key("nodes", 'farms', 'countries', 'totalCru', 'totalSru',
                                                     'totalMru', 'totalHru', 'publicIps', 'accessNodes', 'gateways', 'twins', 'contracts', 'nodesDistribution')
    assertpy.assert_that(len(response_text['nodesDistribution'])).is_greater_than_or_equal_to(1)


def test_stats_qanet():
    response = requests.get(Qanet_URL+'stats?status=up')
    response_text = response.json()

    assertpy.assert_that(response.status_code).is_equal_to(200)
    assertpy.assert_that(response_text).contains_key("nodes", 'farms', 'countries', 'totalCru', 'totalSru',
                                                     'totalMru', 'totalHru', 'publicIps', 'accessNodes', 'gateways', 'twins', 'contracts', 'nodesDistribution')
    assertpy.assert_that(len(response_text['nodesDistribution'])).is_greater_than_or_equal_to(1)

    response = requests.get(Qanet_URL+'stats?status=down')
    response_text = response.json()

    assertpy.assert_that(response.status_code).is_equal_to(200)
    assertpy.assert_that(response_text).contains_key("nodes", 'farms', 'countries', 'totalCru', 'totalSru',
                                                     'totalMru', 'totalHru', 'publicIps', 'accessNodes', 'gateways', 'twins', 'contracts', 'nodesDistribution')
    assertpy.assert_that(len(response_text['nodesDistribution'])).is_greater_than_or_equal_to(1)


def test_stats_testnet():
    response = requests.get(Testnet_URL+'stats?status=up')
    response_text = response.json()

    assertpy.assert_that(response.status_code).is_equal_to(200)
    assertpy.assert_that(response_text).contains_key("nodes", 'farms', 'countries', 'totalCru', 'totalSru',
                                                     'totalMru', 'totalHru', 'publicIps', 'accessNodes', 'gateways', 'twins', 'contracts', 'nodesDistribution')
    assertpy.assert_that(len(response_text['nodesDistribution'])).is_greater_than_or_equal_to(1)

    response = requests.get(Testnet_URL+'stats?status=down')
    response_text = response.json()

    assertpy.assert_that(response.status_code).is_equal_to(200)
    assertpy.assert_that(response_text).contains_key("nodes", 'farms', 'countries', 'totalCru', 'totalSru',
                                                     'totalMru', 'totalHru', 'publicIps', 'accessNodes', 'gateways', 'twins', 'contracts', 'nodesDistribution')
    assertpy.assert_that(len(response_text['nodesDistribution'])).is_greater_than_or_equal_to(1)


def test_stats_mainnet():
    response = requests.get(Mainnet_URL+'stats?status=up')
    response_text = response.json()

    assertpy.assert_that(response.status_code).is_equal_to(200)
    assertpy.assert_that(response_text).contains_key("nodes", 'farms', 'countries', 'totalCru', 'totalSru',
                                                     'totalMru', 'totalHru', 'publicIps', 'accessNodes', 'gateways', 'twins', 'contracts', 'nodesDistribution')
    assertpy.assert_that(len(response_text['nodesDistribution'])).is_greater_than_or_equal_to(1)

    response = requests.get(Mainnet_URL+'stats?status=down')
    response_text = response.json()

    assertpy.assert_that(response.status_code).is_equal_to(200)
    assertpy.assert_that(response_text).contains_key("nodes", 'farms', 'countries', 'totalCru', 'totalSru',
                                                     'totalMru', 'totalHru', 'publicIps', 'accessNodes', 'gateways', 'twins', 'contracts', 'nodesDistribution')
    assertpy.assert_that(len(response_text['nodesDistribution'])).is_greater_than_or_equal_to(1)