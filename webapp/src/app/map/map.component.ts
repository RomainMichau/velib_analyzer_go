import {Component, OnInit} from '@angular/core';
import {MapService} from './map.service';
import * as L from 'leaflet';
import {Station} from "./station";

@Component({
  selector: 'app-map',
  templateUrl: './map.component.html',
  styleUrls: ['./map.component.css']
})
export class MapComponent implements OnInit {
  // @ts-ignore
  private map;
  private markers: L.Layer[] = [];
  private greenIcon = L.icon({
    iconUrl: 'https://www.freeiconspng.com/thumbs/human-icon-png/download-link-for-eps--svg-or-file--0.png',

    iconSize:     [38, 95], // size of the icon
    shadowSize:   [50, 64], // size of the shadow
    iconAnchor:   [22, 94], // point of the icon which will correspond to marker's location
    shadowAnchor: [4, 62],  // the same for the shadow
    popupAnchor:  [-3, -76] // point from which the popup should open relative to the iconAnchor
  });

  constructor(private mapService: MapService) {
  }


  getDow(): number {
    const d = new Date();
    let day = d.getUTCDay()
    if (day == 0) {
      return 7
    }
    return day
  }

  ngOnInit() {
    this.map = L.map('map').setView([0, 0], 13);
    navigator.geolocation.getCurrentPosition((position) => {
      let dow = this.getDow()
      let lat = position.coords.latitude
      let long = position.coords.longitude
      this.map.setView([lat, long], 13);
      L.marker([lat, long], {icon: this.greenIcon}).addTo(this.map);
    L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
      attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
    }).addTo(this.map);
    this.mapService.getCoordinates(lat, long, dow).subscribe(coordinates => {
      coordinates.forEach(station => {
        let coordinates: [number, number] = [station.Latitude, station.Longitude]
        const marker = L.marker(coordinates).addTo(this.map);
        marker.bindPopup(this.toStringA(station, dow)).openPopup();
        this.markers.push(marker);
      });
    })
  })};

  toStringA(station: Station, dow: number): string {
    const arrivals = station.Arrival
    let d = new Date();
    let currentHour = d.getUTCHours();
    let startHour = currentHour - 2;
    let endHour = currentHour + 2;
    let res = `<a href="https://www.google.com/maps/search/?api=1&query=${station.Latitude},${station.Longitude}">Google Maps</a> </br>`
    for (var hour in arrivals[dow]) {
      let hourNumber = Number(hour);
      if (hourNumber ==  currentHour) {
        res += `<b>${hour}h: ${arrivals[dow][hour]} bike/h</b></br>`
      }
      if (hourNumber > startHour && hourNumber < endHour) {
        res += `${hour}h: ${arrivals[dow][hour]} bike/h</br>`
      }
    }
    return res
  }
}
