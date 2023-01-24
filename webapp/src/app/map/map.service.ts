import {Injectable} from '@angular/core';
import {HttpClient, HttpParams} from '@angular/common/http';
import {map, Observable} from 'rxjs';
import {Station} from "./station";

@Injectable({
  providedIn: 'root'
})


export class MapService {
  private coordinatesUrl = 'api/by_dist';

  constructor(private http: HttpClient) {
  }

  convertArrivalTimezoneAndRound(station: Station, offset: number): Station {
    let arrivalWithOffset: { [dow: number]: { [hour: number]: number } } = {}
    for (let dow in station.Arrival) {
      arrivalWithOffset[dow] = {}
      for (let hour in station.Arrival[dow]) {
        let hourNumber = Number(hour);
        arrivalWithOffset[dow][(hourNumber + offset) % 24] = Math.round(station.Arrival[dow][hour] * 100) / 100
      }
    }
    station.Arrival = arrivalWithOffset
    return station;
  }


  getCoordinates(lat: number, long: number, day: number, dist: number): Observable<Array<Station>> {
    let params = new HttpParams().appendAll({"long": long, "lat": lat, dist: dist, dow: day})
    // return of([[48.834882358514875, 2.3045250711792886]]);
    let offset = Math.trunc(new Date().getTimezoneOffset() / -60);
    return this.http.get<Array<Station>>(this.coordinatesUrl, {params: params}).pipe(
      map(stations => {
        return stations.map(station => this.convertArrivalTimezoneAndRound(station, offset));
      }))
  }
}
