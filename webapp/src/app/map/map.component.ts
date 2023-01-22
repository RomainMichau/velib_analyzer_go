import {Component, OnInit} from '@angular/core';
import {MapService} from './map.service';
import * as L from 'leaflet';

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

  ngOnInit() {
    this.map = L.map('map').setView([0, 0], 13);
    navigator.geolocation.getCurrentPosition((position) => {
      let lat = position.coords.latitude
      let long = position.coords.longitude
      this.map.setView([lat, long], 13);
      L.marker([lat, long], {icon: this.greenIcon}).addTo(this.map);
    L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
      attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
    }).addTo(this.map);
    this.mapService.getCoordinates(lat, long).subscribe(coordinates => {
      coordinates.forEach(station => {
        let coordinates: [number, number] = [station.Latitude, station.Longitude]
        const marker = L.marker(coordinates).addTo(this.map);
        marker.bindPopup(station.Name).openPopup();
        this.markers.push(marker);
      });
    })
  })};
}
