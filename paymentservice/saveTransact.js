const Payment = require("./model/payment");



async function saveTransact(cardNumber, cardType, amount) {



    payment = new Payment({
        cardNumber,
        cardType,
        amount
    });

    payment.save();
    console.log(" Transaction Saved in database ");

}

module.exports = saveTransact;