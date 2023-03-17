import { OfferModel } from "../models/offer-model";

export const getSd = async (localDescription: OfferModel): Promise<OfferModel| void> => {
    const resp = await fetch('http://127.0.0.1:8080/api/stream', {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify(localDescription),
    });
    const offer = await resp.json();
    return offer;
};