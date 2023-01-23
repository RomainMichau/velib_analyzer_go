import { Injectable } from '@angular/core';
import {HttpClient, HttpParams} from '@angular/common/http';
import { Observable } from 'rxjs';
import { of } from 'rxjs';
import {Station} from "./station";
@Injectable({
  providedIn: 'root'
})



export class MapService {
  private coordinatesUrl = 'api/by_dist';
  constructor(private http: HttpClient) { }

  getCoordinates(lat: number, long: number): Observable<Array<Station>> {
    const d = new Date();
    let day = d.getDay()
    day = day+1
    let params = new HttpParams().appendAll({"long": long, "lat": lat, dist: 1000, dow: day})
    // return of([[48.834882358514875, 2.3045250711792886]]);
    return this.http.get<Array<Station>>(this.coordinatesUrl, {params: params});
  }
}
