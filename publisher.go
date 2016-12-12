package conejo

import (
	"github.com/streadway/amqp"
	"log"
)

func Publish(conn *amqp.Connection, queue Queue, exchange Exchange, body string) error {

	channel, err := createChannel(conn)
	if err != nil {
		return err
	}
	channel.Confirm(false)
	ack, nack := channel.NotifyConfirm(make(chan uint64, 1), make(chan uint64, 1))
	defer channel.Close()

	err = declareExchange(exchange, channel)
	if err != nil {
		return err
	} else {

		err := declareQueue(queue, channel)
		if err != nil {
			return err
		} else {

			err = channel.QueueBind(
				queue.Name,    // queue name
				queue.Name,    // @TODO - FIX ME!!!
				exchange.Name, // exchange
				false,
				nil,
			)
			if err != nil {
				return err
			} else {

				if err = channel.Publish(
					exchange.Name, // publish to an exchange
					queue.Name,    // routing to 0 or more queues
					false,         // mandatory
					false,         // immediate
					amqp.Publishing{
						Headers:         amqp.Table{},
						ContentType:     "text/plain",
						ContentEncoding: "",
						Body:            []byte(body),
						DeliveryMode:    amqp.Transient, // 1=non-persistent, 2=persistent
						Priority:        0,              // 0-9
					},
				); err != nil {
					return err
				} else {

					select {
					case tag := <-ack:
						log.Println("Acked ", tag)
					case tag := <-nack:
						log.Println("Nack alert! ", tag)
					}
					defer channel.Close()
				}
			}
		}
	}
	return nil
}
