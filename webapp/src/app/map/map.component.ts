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
  private userIcon = L.icon({
    iconUrl: 'https://cdn.shopify.com/s/files/1/1061/1924/products/Very_Angry_Emoji_7f7bb8df-d9dc-4cda-b79f-5453e764d4ea_large.png?v=1571606036',

    iconSize: [38, 50], // size of the icon
  });
  dist = 1000;
  private greenIcon = new L.Icon({
    iconUrl: 'https://raw.githubusercontent.com/pointhi/leaflet-color-markers/master/img/marker-icon-2x-green.png',
    shadowUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/0.7.7/images/marker-shadow.png',
    iconSize: [25, 41],
    iconAnchor: [12, 41],
    popupAnchor: [1, -34],
    shadowSize: [41, 41],
  });

  private orangeIcon = new L.Icon({
    iconUrl: 'https://raw.githubusercontent.com/pointhi/leaflet-color-markers/master/img/marker-icon-2x-orange.png',
    shadowUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/0.7.7/images/marker-shadow.png',
    iconSize: [25, 41],
    iconAnchor: [12, 41],
    popupAnchor: [1, -34],
    shadowSize: [41, 41]
  });
  private redIcon = new L.Icon({
    iconUrl: 'https://raw.githubusercontent.com/pointhi/leaflet-color-markers/master/img/marker-icon-2x-red.png',
    shadowUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/0.7.7/images/marker-shadow.png',
    iconSize: [25, 41],
    iconAnchor: [12, 41],
    popupAnchor: [1, -34],
    shadowSize: [41, 41]
  });



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

  toStringA(station: Station, dow: number): {txt: string, nbBikes: number} {
    const arrivals = station.Arrival
    let currentHour = new Date().getHours();
    let startHour = currentHour - 2;
    let endHour = currentHour + 2;
    let nbBikes = 0;
    let txt = `<a href="https://www.google.com/maps/search/?api=1&query=${station.Latitude},${station.Longitude}">Google Maps</a> </br>`
    for (var hour in arrivals[dow]) {
      let hourNumber = Number(hour);
      if (hourNumber == currentHour) {
        nbBikes = arrivals[dow][hour];
        txt += `<b>${hour}h: ${arrivals[dow][hour]} bike/h</b></br>`
      }
      if (hourNumber > startHour && hourNumber < endHour) {
        txt += `${hour}h: ${arrivals[dow][hour]} bike/h</br>`
      }
    }
    return {txt, nbBikes}
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
    this.map.setView([this.lat, this.long])
    if (this.currentPostMarker != undefined) {
      this.map.removeLayer(this.currentPostMarker);
    }
    this.currentPostMarker = L.marker([this.lat, this.long], {icon: this.userIcon}).addTo(this.map);
    let dow = this.getDow()
    this.markers.forEach(marker => {
      this.map.removeLayer(marker);
    });
    if (this.radius != undefined) {
      this.map.removeLayer(this.radius)
    }
    this.radius = L.circle([this.lat, this.long], {radius: this.dist}).addTo(this.map);
    this.mapService.getStationCoordinates(this.lat, this.long, dow, this.dist).subscribe(coordinates => {
      if(coordinates.length == 0) {
        this.radius?.setStyle({fillColor : 'red'})
      } else {
        this.radius?.setStyle({fillColor : 'green'})
      }
      coordinates.forEach(station => {
        let coordinates: [number, number] = [station.Latitude, station.Longitude]
        const stationDetails = this.toStringA(station, dow)
        let icon = this.redIcon;
        if(stationDetails.nbBikes >= 7) {
          icon = this.greenIcon
        } else if(stationDetails.nbBikes >= 5) {
          icon = this.orangeIcon
        }
        const marker = L.marker(coordinates, {icon: icon}).addTo(this.map);
        marker.bindPopup(stationDetails.txt).openPopup();
        this.markers.push(marker);
      });
    });


  }
}
