import asyncio
import aiohttp
import websockets
import json
import logging
from datetime import datetime
from typing import Dict, Optional
# Define your JSON payloads as python dictionaries
api_endpoint = "http://localhost:9999/api/v1/run/request-service"
ws_endpoint = "ws://localhost:9999/api/v1/run/deploy"


def create_logger() -> logging.Logger:
    """
    This function creates and returns a logger with the specified format.

    :return: logger object
    :rtype: logging.Logger
    """
    logger = logging.getLogger('dms_test')
    try:
        handler = logging.FileHandler(f'{datetime.now().strftime("%Y-%m-%d_%H:%M:%S")}_dms_test.log')
    except:
        handler = logging.FileHandler(f'{datetime.now().strftime("%Y-%m-%d_%H%M%S")}_dms_test.log')
    formatter = logging.Formatter('%(asctime)s - %(name)s - %(levelname)s - %(message)s')
    handler.setFormatter(formatter)
    logger.addHandler(handler)
    logger.setLevel(logging.INFO)
    return logger


def get_user_address() -> str:
    """
    This function prompts the user to enter their wallet address and returns it.

    :return: user's wallet address
    :rtype: str
    """
    return input("Please enter your wallet address: ")


def get_request_payload(user_address: str) -> Dict:
    """
    This function creates and returns a request payload based on the given wallet address.

    :param user_address: user's wallet address
    :type user_address: str
    :return: request payload
    :rtype: dict
    """
    return {
        "address_user": user_address,
        "max_ntx": 1,
        "blockchain": "Cardano",
        "service_type": "ml-training-gpu",
        "params": {
            "image_id": "registry.gitlab.com/nunet/ml-on-gpu/ml-on-gpu-service/develop/tensorflow",
            "model_url": "https://raw.githubusercontent.com/PacktPublishing/Hands-On-GPU-Computing-with-Python/master/Chapter10/10.10.%20Writing%20your%20first%20GPU%20accelerated%20machine%20learning%20programs/pytorch_cifar-10.py",
            "packages": []
        },
        "constraints": {
            "CPU": 500,
            "RAM": 2000,
            "VRAM": 2000,
            "power": 170,
            "complexity": "Low",
            "time": 1
        }
    }


def get_status_success() -> Dict:
    """
    This function returns the status for a successful transaction.

    :return: status for a successful transaction
    :rtype: dict
    """
    return {
        "message": {
            "transaction_status": "success",
            "transaction_type": "fund"
        },
        "action": "send-status"
    }


def get_status_failure() -> Dict:
    """
    This function returns the status for a failed transaction.

    :return: status for a failed transaction
    :rtype: dict
    """
    return {
        "message": {
            "transaction_status": "error",
            "transaction_type": "fund"
        },
        "action": "send-status"
    }


def print_and_log(data: Dict, message: str, logger: logging.Logger, is_error: bool=False) -> None:
    """
    This function logs and prints the given data and message.

    :param data: data to print and log
    :type data: dict
    :param message: message to print and log
    :type message: str
    :param logger: logger to use
    :type logger: logging.Logger
    :param is_error: whether the message is an error message, defaults to False
    :type is_error: bool, optional
    """
    print("----------JSON Response----------")
    print(json.dumps(data, indent=4, sort_keys=True))
    print("---------------------------------")
    if is_error:
        logger.error(message)
    else:
        logger.info(message)


async def process_response(ws: websockets.WebSocketClientProtocol, response: aiohttp.ClientResponse, request_payload: Dict, status_success: Dict, status_failure: Dict, logger: logging.Logger) -> Optional[Dict]:
    """
    This function processes a response and returns the response data if the request is successful.
    :param ws: websocket to send status to
    :type ws: websockets.WebSocketClientProtocol
    :param response: response to process
    :type response: aiohttp.ClientResponse
    :param request_payload: request payload that resulted in the response
    :type request_payload: dict
    :param status_success: status for a successful request
    :type status_success: dict
    :param status_failure: status for a failed request
    :type status_failure: dict
    :param logger: logger to use
    :type logger: logging.Logger
    :return: response data if the request is successful, else None
    :rtype: dict, optional
    """
    if response.status == 200:
        data = await response.json()
        if data.get("compute_provider_addr"):
            print_and_log(data, "Request successful", logger)
            await ws.send(json.dumps(status_success))
            return data
        else:
            print_and_log(data, "Request failed", logger, is_error=True)
            await ws.send(json.dumps(status_failure))
    else:
        print ('---------------------------------')
        print(response)
        print ('---------------------------------')
        logger.error("Error making request")
        print("Error making request")
        await ws.send(json.dumps(status_failure))

async def listen_to_websocket(ws: websockets.WebSocketClientProtocol, logger: logging.Logger) -> bool:
    """
    This function listens to a websocket for job status updates and returns whether the job was completed successfully.

    :param ws: websocket to listen to
    :type ws: websockets.WebSocketClientProtocol
    :param logger: logger to use
    :type logger: logging.Logger
    :return: whether the job was completed successfully
    :rtype: bool
    """
    while True:
        try:
            message = await asyncio.wait_for(ws.recv(), timeout=20.0)
            data = json.loads(message)
            print_and_log(data, f"Received: {data}", logger)

            if data.get('action') == "job-completed":
                logger.info("Test successful")
                print("Test successful")
                return True
            elif data.get('action') == "job-failed":
                logger.error("Test failed")
                print("Test failed")
                return False
        except asyncio.TimeoutError:
            logger.error("Timeout while waiting for job status")
            print("Timeout while waiting for job status")
            return False

async def test_simulation() -> None:
    """
    This function runs a test simulation. It asks the user for their wallet address, sends a service request, listens 
    for updates on a websocket, and asks the user if they want to run another test after each simulation.
    """
    logger = create_logger()
    while True:  # Run the simulation in a loop
        user_address = get_user_address()
        request_payload = get_request_payload(user_address)
        status_success = get_status_success()
        status_failure = get_status_failure()

        async with aiohttp.ClientSession() as session:
            try:
                async with websockets.connect(ws_endpoint) as ws:
                    connected_msg = await ws.recv()
                    connected_data = json.loads(connected_msg)
                    if connected_data.get("action") == "connected to mock DMS":
                        print_and_log(connected_data, "Connection successful", logger)
                    else:
                        print_and_log(connected_data, "Connection failed", logger, is_error=True)
                        return

                    async with session.post(api_endpoint, json=request_payload) as resp:
                        data = await process_response(ws, resp, request_payload, status_success, status_failure, logger)
                        if data:
                            success = await listen_to_websocket(ws, logger)
                            if success:
                                logger.info("Simulation complete")
                                print("Simulation complete")
                            else:
                                logger.error("Simulation failed")
                                print("Simulation failed")

            except Exception as e:
                logger.error("An error occurred during the simulation")
                print("An error occurred during the simulation")
                logger.error(str(e))
                print(str(e))

        run_again = input("Do you want to run another test? (yes/no): ")
        if run_again.lower() != "yes":
            break  # End the loop if the user doesn't want to run another test


if __name__ == "__main__":
    asyncio.run(test_simulation())


