/**
 * Copyright 2022 Google LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import { DataSourceInstanceSettings, QueryFixAction, ScopedVars } from '@grafana/data';
import { DataSourceWithBackend, getTemplateSrv, TemplateSrv } from '@grafana/runtime';
import { CloudLoggingOptions, Query } from './types';
import { CloudLoggingVariableSupport } from './variables';

export class DataSource extends DataSourceWithBackend<Query, CloudLoggingOptions> {
  authenticationType: string;
  annotations = {};

  constructor(
    private instanceSettings: DataSourceInstanceSettings<CloudLoggingOptions>,
    private readonly templateSrv: TemplateSrv = getTemplateSrv(),
  ) {
    super(instanceSettings);
    this.authenticationType = instanceSettings.jsonData.authenticationType || 'jwt';
    this.variables = new CloudLoggingVariableSupport(this);
  }

  /**
   * Get the Project ID from GCE or we parsed from the data source's JWT token
   *
   * @returns Project ID
   */
  async getDefaultProject() {
    try {
      const { defaultProject, authenticationType } = this.instanceSettings.jsonData;
      if (authenticationType === 'gce') {
        await this.ensureGCEDefaultProject();
        return this.instanceSettings.jsonData.gceDefaultProject || "";
      }

      return defaultProject || '';
    } catch (error) {
      console.error('Failed to get default project:', error);
      return '';
    }
  }

  async getGCEDefaultProject() {
    return this.getResource(`gceDefaultProject`);
  }

  async ensureGCEDefaultProject() {
    try {
      const { authenticationType, gceDefaultProject } = this.instanceSettings.jsonData;
      if (authenticationType === 'gce' && !gceDefaultProject) {
        this.instanceSettings.jsonData.gceDefaultProject = await this.getGCEDefaultProject();
      }
    } catch (error) {
      console.error('Failed to ensure GCE default project:', error);
      throw error;
    }
  }

  /**
   * Have the backend call `resourcemanager.projects.list` with our credentials,
   * and return the IDs of all projects found
   *
   * @returns List of discovered project IDs
   */
  async getProjects(): Promise<string[]> {
    try {
      return await this.getResource(`projects`);
    } catch (error) {
      console.error('Failed to fetch projects:', error);
      return [];
    }
  }

  /**
   * Have the backend call `projects.locations.buckets.list` with our credentials,
   * and return the names of all log buckets found
   *
   * @returns List of discovered bucket names
   */
  async getLogBuckets(projectId: string): Promise<string[]> {
    if (!projectId) {
      console.warn('getLogBuckets called without projectId');
      return [];
    }
    try {
      return await this.getResource(`logBuckets`, { "ProjectId": projectId });
    } catch (error) {
      console.error('Failed to fetch log buckets:', error);
      return [];
    }
  }

  /**
   * Have the backend call `projects.locations.buckets.views.list` with our credentials,
   * and return the names of all views of the log bucket
   *
   * @returns List of discovered bucket names
   */
  async getLogBucketViews(projectId: string, bucketId: string): Promise<string[]> {
    if (!projectId || !bucketId) {
      console.warn('getLogBucketViews called without projectId or bucketId');
      return [];
    }
    try {
      return await this.getResource(`logViews`, { "ProjectId": projectId, "BucketId": bucketId });
    } catch (error) {
      console.error('Failed to fetch log bucket views:', error);
      return [];
    }
  }

  applyTemplateVariables(query: Query, scopedVars: ScopedVars): Query {
    return {
      ...query,
      queryText: this.templateSrv.replace(query.queryText, scopedVars),
      projectId: this.templateSrv.replace(query.projectId, scopedVars),
      bucketId: this.templateSrv.replace(query.bucketId, scopedVars),
      viewId: this.templateSrv.replace(query.viewId, scopedVars),
    };
  }

  modifyQuery(query: Query, action: QueryFixAction): Query {
    let queryText = query.queryText;

    switch (action.type) {
      case 'ADD_FILTER': {
        if (action.options?.key && action.options?.value) {
          if (action.options?.key === "id") {
            queryText += `\ninsertId="${escapeLabelValue(action.options.value)}"`;
          } else if (action.options?.key === "level") {
            let level = action.options.value;
            if (level === "debug") {
              level = "DEFAULT";
            } else if (level === "critical") {
              level = "EMERGENCY";
            }
            queryText += `\nseverity="${escapeLabelValue(level)}"`;
          } else {
            queryText += `\n${action.options.key}="${escapeLabelValue(action.options.value)}"`;
          }
        }
        break;
      }
      case 'ADD_FILTER_OUT': {
        if (action.options?.key && action.options?.value) {
          if (action.options?.key === "id") {
            queryText += `\ninsertId!="${escapeLabelValue(action.options.value)}"`;
          } else if (action.options?.key === "level") {
            let level = action.options.value;
            if (level === "debug") {
              level = "DEFAULT";
            } else if (level === "critical") {
              level = "EMERGENCY";
            }
            queryText += `\nseverity!="${escapeLabelValue(level)}"`;
          } else {
            queryText += `\n${action.options.key}!="${escapeLabelValue(action.options.value)}"`;
          }
        }
        break;
      }
    }
    return { ...query, queryText: queryText };
  }

  filterQuery(query: Query): boolean {
    return !query.hide;
  }
}

// the 3 symbols we handle are:
// - \n ... the newline character
// - \  ... the backslash character
// - "  ... the double-quote character
function escapeLabelValue(labelValue: string): string {
  return labelValue.replace(/\\/g, '\\\\').replace(/\n/g, '\\n').replace(/"/g, '\\"');
}
