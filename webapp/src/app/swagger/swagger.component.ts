import {Component} from '@angular/core';
import {SwaggerUIBundle} from "swagger-ui-dist";
@Component({
  selector: 'swagger',
  templateUrl: './swagger.component.html',
})
export class SwaggerComponent  {
  ngAfterViewInit() {
    SwaggerUIBundle({
      url: "swagger.json",
      showMutatedRequest: true,
      // disable the validator thingy that's always red
      validatorUrl: null,
      domNode: document.getElementById('swagger-ui-item'),
      presets: [
        SwaggerUIBundle.presets.apis
      ],
      plugins: [
        SwaggerUIBundle.plugins.DownloadUrl
      ],
      defaultModelRendering: 'model',
      defaultModelExpandDepth: 3,
      syntaxHighlight: {
        activated: true
      },
      filter: true,
    });
  }
}
