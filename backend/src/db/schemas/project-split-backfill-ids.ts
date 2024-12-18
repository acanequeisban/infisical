// Code generated by automation script, DO NOT EDIT.
// Automated by pulling database and generating zod schema
// To update. Just run npm run generate:schema
// Written by akhilmhdh.

import { z } from "zod";

import { TImmutableDBKeys } from "./models";

export const ProjectSplitBackfillIdsSchema = z.object({
  id: z.string().uuid(),
  sourceProjectId: z.string(),
  destinationProjectType: z.string(),
  destinationProjectId: z.string()
});

export type TProjectSplitBackfillIds = z.infer<typeof ProjectSplitBackfillIdsSchema>;
export type TProjectSplitBackfillIdsInsert = Omit<z.input<typeof ProjectSplitBackfillIdsSchema>, TImmutableDBKeys>;
export type TProjectSplitBackfillIdsUpdate = Partial<
  Omit<z.input<typeof ProjectSplitBackfillIdsSchema>, TImmutableDBKeys>
>;