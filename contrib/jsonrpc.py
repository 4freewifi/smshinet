import urllib2
import json


serial = 0


def rpc_call(url, method, args):
    global serial
    data = json.dumps({
        'jsonrpc': '2.0',
        'method': method,
        'params': args,
        'id': serial,
    })
    serial += 1
    req = urllib2.Request(
        url, data, {'Content-Type': 'application/json'})
    f = urllib2.urlopen(req)
    res = f.read()
    return json.loads(res)


def main():
    url = 'http://localhost:3059/jsonrpc'
    try:
        print rpc_call(url, 'Echo.Echo', {'in': 'TEST'})
        print rpc_call(url, 'SMSHiNet.SendTextSMS', {
            'recipient': '0912345678',
            'message': 'jsonrpc test',
        })
    except urllib2.HTTPError as e:
        print e.code, e.read()


if __name__ == '__main__':
    main()
