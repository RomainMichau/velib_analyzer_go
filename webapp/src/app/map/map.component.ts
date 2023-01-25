import {Component, OnInit} from '@angular/core';
import {MapService} from './map.service';
import * as L from 'leaflet';
import {Circle} from 'leaflet';
import {Station} from "./station";
import {ActivatedRoute, Router} from "@angular/router";

@Component({
  selector: 'app-map',
  templateUrl: './map.component.html',
  styleUrls: ['./map.component.css']
})
export class MapComponent implements OnInit {
  // @ts-ignore
  private map;
  private radius: Circle | undefined
  private lat: number = 0
  private long: number = 0
  private currentPostMarker: L.Layer | undefined
  private markers: L.Layer[] = [];
  private greenIcon = L.icon({
    iconUrl: 'https://www.freeiconspng.com/thumbs/human-icon-png/download-link-for-eps--svg-or-file--0.png',

    iconSize: [38, 95], // size of the icon
    shadowSize: [50, 64], // size of the shadow
    iconAnchor: [22, 94], // point of the icon which will correspond to marker's location
    shadowAnchor: [4, 62],  // the same for the shadow
    popupAnchor: [-3, -76] // point from which the popup should open relative to the iconAnchor
  });
  dist = 1000;

  constructor(private mapService: MapService, private route: ActivatedRoute, private router: Router) {
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
    this.route.queryParams.subscribe(params => {
      this.dist = params['dist'] || this.dist;
    });
    this.map = L.map('map');
    this.map.on('click', (e: any) => {
      this.lat = e.latlng.lat;
      this.long = e.latlng.lng;
      this.updateMap();
    });
    navigator.geolocation.getCurrentPosition((position) => {
      this.lat = position.coords.latitude
      this.long = position.coords.longitude
      this.map.setView([this.lat, this.long], 13);
      L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
        attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
      }).addTo(this.map);
      this.updateMap()
    })
  };

  toStringA(station: Station, dow: number): string {
    const arrivals = station.Arrival
    let currentHour = new Date().getHours();
    let startHour = currentHour - 2;
    let endHour = currentHour + 2;
    let res = `<a href="https://www.google.com/maps/search/?api=1&query=${station.Latitude},${station.Longitude}">Google Maps</a> </br>`
    for (var hour in arrivals[dow]) {
      let hourNumber = Number(hour);
      if (hourNumber == currentHour) {
        res += `<b>${hour}h: ${arrivals[dow][hour]} bike/h</b></br>`
      }
      if (hourNumber > startHour && hourNumber < endHour) {
        res += `${hour}h: ${arrivals[dow][hour]} bike/h</br>`
      }
    }
    return res
  }

  resetLocation() {
    navigator.geolocation.getCurrentPosition((position) => {
      this.lat = position.coords.latitude;
      this.long = position.coords.longitude;
      this.updateMap();
    });
  }

  updateMap() {
    this.router.navigate([], {queryParams: {dist: this.dist, lat: this.lat, long: this.long }, relativeTo: this.route});
    if (this.currentPostMarker != undefined) {
      this.map.removeLayer(this.currentPostMarker);
    }
    this.currentPostMarker = L.marker([this.lat, this.long], {icon: this.greenIcon}).addTo(this.map);
    let dow = this.getDow()
    this.mapService.getCoordinates(this.lat, this.long, dow, this.dist).subscribe(coordinates => {
      this.markers.forEach(marker => {
        this.map.removeLayer(marker);
      });
      coordinates.forEach(station => {
        let coordinates: [number, number] = [station.Latitude, station.Longitude]
        const marker = L.marker(coordinates).addTo(this.map);
        marker.bindPopup(this.toStringA(station, dow)).openPopup();
        this.markers.push(marker);
      });
    });
    if (this.radius != undefined) {
      this.map.removeLayer(this.radius)
    }
    this.radius = L.circle([this.lat, this.long], {radius: this.dist}).addTo(this.map);

  }
}
