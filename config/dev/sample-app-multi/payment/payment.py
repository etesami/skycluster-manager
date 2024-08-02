from flask import Flask, request, jsonify
import logging

app = Flask(__name__)
logging.basicConfig(level=logging.INFO) 

@app.route('/', methods=['POST'])
def process_payment():
    # Dummy payment processing logic
    data = request.get_json()
    amount = data.get('amount')
    app.logger.info('Payment processing request received. Amount:', amount)
    # Dummy payment validation
    if amount:
        return jsonify({'status': 'success', 'message': 'Payment processed successfully. Amount: $' + str(amount)})
    else:
        return jsonify({'status': 'error', 'message': 'Invalid payment data'})

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000)



# curl -X POST -H "Content-Type: application/json" -d '{"amount": 120}' http://payment:5000