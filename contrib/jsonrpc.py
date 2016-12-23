# python3

import argparse
import json
import time

import requests



serial = 0


def rpc_call(url, method, args):
    global serial
    data = {
        'jsonrpc': '2.0',
        'method': method,
        'params': args,
        'id': serial,
    }
    serial += 1
    r = requests.post(url, json=data)
    return r.json()


def main():
    p = argparse.ArgumentParser(description='Command line tool for smshinetd')
    p.add_argument('--uri', '-u', default='http://localhost:3059/jsonrpc',
                   help='URI of smshinetd service')
    p.add_argument('--dry-run', '-n', action='store_true')
    p.add_argument('recipient', help='Mobile number of recipient')
    p.add_argument('message', nargs='?', default='jsonrpc test',
                   help='Message to send')
    args = p.parse_args()
    print(rpc_call(args.uri, 'Echo.Echo', {'in': 'TEST'}))
    if args.dry_run:
        return
    r = rpc_call(args.uri, 'SMSHiNet.SendTextSMS', {
        'recipient': args.recipient,
        'message': args.message,
    })
    print(r, ', wait for 30 secs to check')
    msgid = r['result']
    time.sleep(30)
    r = rpc_call(args.uri, 'SMSHiNet.CheckTextStatus', msgid)
    print(r)


if __name__ == '__main__':
    main()
